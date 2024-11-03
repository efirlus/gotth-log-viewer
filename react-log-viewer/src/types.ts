export interface LogEntry {
    timestamp: string;
    level: 'error' | 'warn' | 'info' | 'debug';
    program: string;
    message: string;
    location?: string;  // Optional location field
    raw?: any;         // Store the original log entry
  }

export interface LogFilters {
  program: string | null;
  search: string;
  level: string | null;
}