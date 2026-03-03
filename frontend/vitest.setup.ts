import '@testing-library/jest-dom/vitest';

Object.defineProperty(window, 'confirm', {
	writable: true,
	value: vi.fn(() => true)
});

Object.defineProperty(window.navigator, 'clipboard', {
	writable: true,
	value: {
		writeText: vi.fn().mockResolvedValue(undefined)
	}
});
