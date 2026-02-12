/**
 * WebSocket Store for Real-Time Collaboration
 *
 * Manages WebSocket connections for project-based real-time collaboration.
 * Features:
 * - Automatic reconnection with exponential backoff
 * - User presence tracking
 * - Bulk operation sync
 */

export interface UserPresence {
	user_id: string;
	first_name: string;
	last_name: string;
	email: string;
}

interface WSState {
	connected: boolean;
	connecting: boolean;
	error: string | null;
	projectId: string | null;
	activeUsers: UserPresence[];
}

class WebSocketStore {
	private ws: WebSocket | null = $state(null);
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

	// Reactive state
	private state = $state<WSState>({
		connected: false,
		connecting: false,
		error: null,
		projectId: null,
		activeUsers: []
	});

	// Getters for reactive access
	get connected() {
		return this.state.connected;
	}

	get connecting() {
		return this.state.connecting;
	}

	get error() {
		return this.state.error;
	}

	get activeUsers() {
		return this.state.activeUsers;
	}

	get projectId() {
		return this.state.projectId;
	}

	// Derived value
	get userCount() {
		return this.state.activeUsers.length;
	}

	/**
	 * Connect to a project's WebSocket room
	 */
	connect(projectId: string) {
		if (this.state.projectId === projectId && this.state.connected) {
			return; // Already connected to this project
		}

		this.disconnect(); // Clean up existing connection

		this.state.projectId = projectId;
		this.state.connecting = true;
		this.state.error = null;

		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${protocol}//${window.location.host}/api/v1/projects/${projectId}/ws`;

		try {
			this.ws = new WebSocket(wsUrl);

			this.ws.onopen = () => {
				this.state.connected = true;
				this.state.connecting = false;
				this.state.error = null;
				this.reconnectAttempts = 0;
				console.log(`[WS] Connected to project ${projectId}`);
			};

			this.ws.onmessage = (event) => {
				try {
					const message = JSON.parse(event.data);
					this.handleMessage(message);
				} catch (err) {
					console.error('[WS] Failed to parse message:', err);
				}
			};

			this.ws.onerror = (error) => {
				console.error('[WS] Error:', error);
				this.state.error = 'Connection error';
			};

			this.ws.onclose = (event) => {
				console.log('[WS] Connection closed:', event.code, event.reason);
				this.state.connected = false;
				this.state.connecting = false;

				// Attempt reconnection if not a clean close
				if (event.code !== 1000) {
					this.attemptReconnect();
				}
			};
		} catch (error) {
			console.error('[WS] Connection failed:', error);
			this.state.connecting = false;
			this.state.error = error instanceof Error ? error.message : 'Connection failed';
		}
	}

	/**
	 * Disconnect from the current WebSocket
	 */
	disconnect() {
		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = null;
		}

		if (this.ws) {
			this.ws.close(1000, 'Client disconnect');
			this.ws = null;
		}

		this.state.connected = false;
		this.state.connecting = false;
		this.state.projectId = null;
		this.state.activeUsers = [];
		this.reconnectAttempts = 0;
	}

	/**
	 * Attempt to reconnect with exponential backoff
	 */
	private attemptReconnect() {
		if (!this.state.projectId || this.reconnectAttempts >= this.maxReconnectAttempts) {
			if (this.reconnectAttempts >= this.maxReconnectAttempts) {
				this.state.error = 'Max reconnection attempts reached';
			}
			return;
		}

		this.reconnectAttempts++;
		const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);

		console.log(
			`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
		);

		this.reconnectTimeout = setTimeout(() => {
			if (this.state.projectId) {
				this.connect(this.state.projectId);
			}
		}, delay);
	}

	/**
	 * Handle incoming WebSocket messages
	 */
	private handleMessage(message: any) {
		console.log('[WS] Message received:', message.action);

		switch (message.action) {
			case 'PRESENCE_LIST':
				this.state.activeUsers = message.payload.users || [];
				console.log('[WS] Presence list updated:', this.state.activeUsers.length, 'users');
				break;

			case 'PRESENCE_JOIN':
				if (message.payload.user) {
					const exists = this.state.activeUsers.some(
						(u) => u.user_id === message.payload.user.user_id
					);
					if (!exists) {
						this.state.activeUsers = [...this.state.activeUsers, message.payload.user];
						console.log('[WS] User joined:', message.payload.user.email);
					}
				}
				break;

			case 'PRESENCE_LEAVE':
				if (message.payload.user_id) {
					this.state.activeUsers = this.state.activeUsers.filter(
						(u) => u.user_id !== message.payload.user_id
					);
					console.log('[WS] User left:', message.payload.user_id);
				}
				break;

			case 'BULK_UPDATE_ELEMENTS':
			case 'BULK_DELETE_ELEMENTS':
			case 'BULK_CREATE_ELEMENTS':
				// Dispatch custom event for components to handle
				window.dispatchEvent(new CustomEvent('ws-bulk-operation', { detail: message }));
				console.log('[WS] Bulk operation:', message.action, message.payload?.entity_type);
				break;

			case 'ERROR':
				console.error('[WS] Server error:', message.payload);
				break;

			default:
				console.warn('[WS] Unknown action:', message.action);
		}
	}

	/**
	 * Send a message via WebSocket
	 */
	send(action: string, payload: any) {
		if (!this.ws || !this.state.connected || !this.state.projectId) {
			throw new Error('WebSocket not connected');
		}

		const message = {
			action,
			payload,
			project_id: this.state.projectId
		};

		this.ws.send(JSON.stringify(message));
		console.log('[WS] Message sent:', action);
	}

	/**
	 * Send a bulk update operation
	 */
	bulkUpdate(entityType: string, updates: any[]) {
		this.send('BULK_UPDATE_ELEMENTS', {
			entity_type: entityType,
			updates
		});
	}

	/**
	 * Send a bulk delete operation
	 */
	bulkDelete(entityType: string, ids: string[]) {
		this.send('BULK_DELETE_ELEMENTS', {
			entity_type: entityType,
			ids
		});
	}

	/**
	 * Send a bulk create operation
	 */
	bulkCreate(entityType: string, items: any[]) {
		this.send('BULK_CREATE_ELEMENTS', {
			entity_type: entityType,
			items
		});
	}
}

// Export singleton instance
export const websocketStore = new WebSocketStore();
