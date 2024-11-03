import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import type { ViteDevServer, Plugin } from 'vite';

// Create a custom plugin for the logs API
const logsApiPlugin: Plugin = {
  name: 'logs-api',
  configureServer(server: ViteDevServer) {
    server.middlewares.use('/api/logs', (_req, res) => {
      try {
        // Read the log file
        const logs = fs.readFileSync('/home/efirlus/goproject/Logs/app.log', 'utf-8');
        console.log('Reading logs from app.log...');
        
        // Parse log entries
        const logEntries = logs
          .split('\n')
          .filter(line => line.trim())
          .map(line => {
            try {
              return JSON.parse(line);
            } catch (e) {
              // Try to parse as plain text if JSON parsing fails
              return {
                timestamp: new Date().toISOString(),
                level: 'info',
                program: 'system',
                message: line
              };
            }
          });

        console.log(`Found ${logEntries.length} log entries`);
        
        const sinceTimestamp = new URL(_req.url!, 'http://localhost').searchParams.get('since')
        const filteredLogs = sinceTimestamp
          ? logEntries.filter(log => {
              const logTime = new Date(log.time || log.timestamp).getTime();
              const sinceTime = new Date(sinceTimestamp).getTime();
              return logTime > sinceTime;
            })
          : logEntries

        console.log(`Found ${filteredLogs.length} log entries${sinceTimestamp ? ' since ' + sinceTimestamp : ''}`);
        
        res.setHeader('Content-Type', 'application/json');
        res.end(JSON.stringify(logEntries));
      } catch (error) {
        console.error('Error reading logs:', error);
        res.statusCode = 500;
        res.end(JSON.stringify({ 
          error: 'Failed to read logs',
          details: error instanceof Error ? error.message : 'Unknown error'
        }));
      }
    });
  }
};

export default defineConfig({
  plugins: [react(), logsApiPlugin],
});