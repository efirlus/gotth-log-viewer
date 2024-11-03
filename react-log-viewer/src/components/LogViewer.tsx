import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { Search, Filter, X, ChevronDown, ChevronRight } from 'lucide-react';
import type { LogEntry, LogFilters } from '../types';
import { fetchLogs, resetLastTimestamp } from '../services/logService';

const POLLING_INTERVAL = 2000;

// Catppuccin Mocha color scheme
const colors = {
    base: 'bg-[#1e1e2e]', // Base background
    mantle: 'bg-[#181825]', // Darker background
    crust: 'bg-[#11111b]', // Darkest background
    text: 'text-[#cdd6f4]', // Text
    subtext: 'text-[#a6adc8]', // Secondary text
    overlay0: 'text-[#6c7086]', // Muted text
    surface0: 'bg-[#313244]', // Surface background
    surface1: 'bg-[#45475a]', // Lighter surface
    blue: 'bg-[#89b4fa]',
    lavender: 'bg-[#b4befe]',
    sapphire: 'bg-[#74c7ec]',
    red: 'bg-[#f38ba8]',
    peach: 'bg-[#fab387]',
    yellow: 'bg-[#f9e2af]',
    green: 'bg-[#a6e3a1]',
    mauve: 'bg-[#cba6f7]'
  };

const LogViewer: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedLogs, setExpandedLogs] = useState<Set<number>>(new Set());
  
  const [filters, setFilters] = useState<LogFilters>({
    program: null,
    search: '',
    level: null,
  });

  const loadLogs = useCallback(async (isInitial: boolean = false) => {
    try {
        if (isInitial) {
            setLoading(true);
            resetLastTimestamp();
        }
        const newLogs = await fetchLogs();
        setLogs(prevLogs => {
            if (isInitial) return newLogs;
            const existingTimestamps = new Set(prevLogs.map(log => log.timestamp));
            const uniqueNewLogs = newLogs.filter(log => !existingTimestamps.has(log.timestamp));
            return [...prevLogs, ...uniqueNewLogs];
        });
        setError(null);
    } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to load logs';
      setError(errorMessage);
      console.error('Error loading logs:', err);
    } finally {
      if (isInitial) setLoading(false);
    
    }
  }, []);

  useEffect(() => {
    loadLogs(true);
  }, [loadLogs]);

  useEffect(() => {
    const pollInterval = setInterval(() => {
        loadLogs(false);
    }, POLLING_INTERVAL);
    return () => clearInterval(pollInterval);
  }, [loadLogs]);

  const uniquePrograms = useMemo(() => {
    return Array.from(new Set(logs.map(log => log.program))).sort();
  }, [logs]);

  const filteredLogs = useMemo(() => {
    return logs.filter(log => {
      const matchesProgram = !filters.program || log.program === filters.program;
      const matchesLevel = !filters.level || log.level === filters.level;
      const matchesSearch = !filters.search || 
        log.message.toLowerCase().includes(filters.search.toLowerCase()) ||
        log.program.toLowerCase().includes(filters.search.toLowerCase()) ||
        (log.location?.toLowerCase().includes(filters.search.toLowerCase()) ?? false);
      
      return matchesProgram && matchesLevel && matchesSearch;
    })
    .sort((a,b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
  }, [logs, filters]);

  const handleProgramClick = (program: string) => {
    setFilters(prev => ({
      ...prev,
      program: prev.program === program ? null : program,
    }));
  };

  const clearFilters = () => {
    setFilters({
      program: null,
      search: '',
      level: null,
    });
  };

  const toggleLogExpansion = (index: number) => {
    setExpandedLogs(prev => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  if (loading) {
    return (
      <div className="min-h-screen ${colors.base} p-6 flex items-center justify-center">
        <div className="text-gray-500">Loading logs...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen ${colors.base} p-6 flex items-center justify-center">
        <div className="text-red-500">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen ${colors.base} p-6`}>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className={`text-3xl font-bold ${colors.text} mb-4`}>Log Viewer</h1>
          
          {/* Filters */}
          <div className={`${colors.mantle} rounded-lg shadow-md p-4 mb-6 border border-opacity-10 ${colors.surface0}`}>
            <div className="flex items-center gap-4 mb-4">
              <div className="flex-1 relative">
                <Search className={`absolute left-3 top-1/2 -translate-y-1/2 ${colors.overlay0} h-5 w-5`} />
                <input
                  type="text"
                  placeholder="Search logs..."
                  className={`w-full pl-10 pr-4 py-2 ${colors.crust} ${colors.text} border border-opacity-10 ${colors.surface0} rounded-lg focus:ring-2 focus:ring-opacity-50 focus:ring-[#89b4fa] focus:border-transparent`}
                  value={filters.search}
                  onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                />
              </div>
              <select
                className={`px-4 py-2 ${colors.crust} ${colors.text} border border-opacity-10 ${colors.surface0} rounded-lg focus:ring-2 focus:ring-opacity-50 focus:ring-[#89b4fa]`}
                value={filters.level || ''}
                onChange={(e) => setFilters(prev => ({ ...prev, level: e.target.value || null }))}
              >
                <option value="">All Levels</option>
                <option value="error">Error</option>
                <option value="warn">Warning</option>
                <option value="info">Info</option>
                <option value="debug">Debug</option>
              </select>
              {(filters.program || filters.search || filters.level) && (
                <button
                  onClick={clearFilters}
                  className={`flex items-center gap-2 px-4 py-2 text-sm ${colors.subtext} hover:${colors.text}`}
                >
                  <X className="h-4 w-4" />
                  Clear filters
                </button>
              )}
            </div>
            
            <div className="flex flex-wrap gap-2">
              {uniquePrograms.map(program => (
                <button
                  key={program}
                  onClick={() => handleProgramClick(program)}
                  className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium transition-colors
                    ${filters.program === program
                      ? `${colors.lavender} text-[#1e1e2e] ring-2 ring-[#b4befe]`
                      : `${colors.surface0} ${colors.subtext} hover:${colors.surface1}`
                    }`}
                >
                  <Filter className="h-4 w-4 mr-1" />
                  {program}
                </button>
              ))}
            </div>
          </div>

          {/* Log Entries */}
          <div className={`${colors.mantle} rounded-lg shadow-md divide-y divide-[#313244]`}>
            {filteredLogs.length === 0 ? (
              <div className={`p-4 text-center ${colors.overlay0}`}>
                No logs found matching your criteria
              </div>
            ) : (
              filteredLogs.map((log, index) => (
                <div key={index} className={`p-4 hover:${colors.surface0} transition-colors`}>
                  <div className="flex items-center gap-3 mb-2">
                    <button
                      onClick={() => toggleLogExpansion(index)}
                      className={`p-1 hover:${colors.surface1} rounded`}
                    >
                      {expandedLogs.has(index) ? (
                        <ChevronDown className={`h-4 w-4 ${colors.overlay0}`} />
                      ) : (
                        <ChevronRight className={`h-4 w-4 ${colors.overlay0}`} />
                      )}
                    </button>
                    <span className={`text-sm ${colors.overlay0} font-mono`}>
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium
                      ${log.level === 'error' ? `${colors.red} text-[#11111b]` :
                        log.level === 'warn' ? `${colors.peach} text-[#11111b]` :
                        log.level === 'info' ? `${colors.yellow} text-[#11111b]` :
                        `${colors.surface0} ${colors.text}`}`}>
                      {log.level.toUpperCase()}
                    </span>
                    <span
                      className={`px-2 py-0.5 rounded-full text-xs font-medium ${colors.mauve} text-[#11111b] cursor-pointer hover:opacity-90`}
                      onClick={() => handleProgramClick(log.program)}
                    >
                      {log.program}
                    </span>
                  </div>
                  <div className="pl-8">
                    <p className={`${colors.text} font-mono whitespace-pre-wrap`}>{log.message}</p>
                    {log.location && expandedLogs.has(index) && (
                      <div className={`mt-2 text-sm ${colors.overlay0} font-mono whitespace-pre-wrap`}>
                        Location:
                        <div className={`pl-4 border-l-2 ${colors.surface0} mt-1`}>
                          {log.location}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>  );
};

export default LogViewer;