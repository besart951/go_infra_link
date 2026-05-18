/**
 * Domain entity representing a Project
 */
import type { Phase } from '../phase/index.js';

export interface Project {
  id: string;
  name: string;
  description: string;
  status: 'planned' | 'ongoing' | 'completed';
  start_date?: string | null;
  phase_id: string;
  phase?: Phase | null;
  creator_id: string;
  created_at: string;
  updated_at: string;
}
