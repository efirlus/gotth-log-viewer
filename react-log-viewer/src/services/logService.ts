import type { LogEntry } from '../types';

let lastTimestamp: string | null = null;

export async function fetchLogs(): Promise<LogEntry[]> {
  try {
    const url = lastTimestamp
      ? `/api/logs?since=${encodeURIComponent(lastTimestamp)}`
      : '/api/logs';

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    
    // Transform the raw log data to match our LogEntry interface
    const logs = data.map((log: any) => ({
      timestamp: log.time || log.timestamp || new Date().toISOString(),
      level: (log.loglevel || log.level || log.severity || 'info').toLowerCase(),
      program: log.programname || log.program || log.service || 'unknown',
      message: log.message || log.msg || log.text || log.logstring || String(log),
      location: log.location || undefined,
      raw: log  // Store the original log entry for debugging
    }));

    if (logs.length > 0) {
        lastTimestamp = logs[logs.length - 1].timestamp
    }

    return logs;
  } catch (error) {
    console.error('Error fetching logs:', error);
    throw error;
  }
}

export function resetLastTimestamp() {
    lastTimestamp = null;
}