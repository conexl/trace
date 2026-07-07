import * as React from 'react';
import { ScrollText, X, Search, Pause, Play, Download } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import type { LogChunk } from '@/lib/types';
import { cn } from '@/lib/utils';

interface LogsPanelProps {
  logs: LogChunk[];
  activeStream: string | null;
  onStreamChange: (name: string | null) => void;
}

export function LogsPanel({ logs, activeStream, onStreamChange }: LogsPanelProps) {
  const scrollRef = React.useRef<HTMLDivElement>(null);
  const [search, setSearch] = React.useState('');
  const [isPaused, setIsPaused] = React.useState(false);
  const streams = React.useMemo(() => {
    const names = new Set<string>();
    logs.forEach((chunk) => names.add(chunk.name));
    return Array.from(names);
  }, [logs]);

  const current = activeStream && streams.includes(activeStream) ? activeStream : streams[0] ?? null;

  const lines = logs
    .filter((chunk) => chunk.name === current)
    .flatMap((chunk) =>
      chunk.data
        .split('\n')
        .filter((line) => line.trim().length > 0)
        .map((line) => ({ line, name: chunk.name, offset: chunk.offset }))
    )
    .filter((l) => !search || l.line.toLowerCase().includes(search.toLowerCase()))
    .slice(-200);

  React.useEffect(() => {
    if (scrollRef.current && !isPaused) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [lines.length, current, isPaused]);

  const handleDownload = () => {
    const content = lines.map(l => l.line).join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${current || 'logs'}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="flex h-full flex-col">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 text-active">
            <ScrollText className="h-4 w-4 text-accent" />
            <span className="text-sm font-medium tracking-tight">Log tail</span>
          </div>
          <div className="relative">
            <Search className="absolute left-2 top-1/2 h-3 w-3 -translate-y-1/2 text-muted" />
            <input
              type="text"
              placeholder="Search logs..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-7 w-40 rounded-md border border-border bg-canvas pl-7 pr-2 text-[10px] text-active focus:border-accent focus:outline-none"
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsPaused(!isPaused)}
            className={cn(
              "flex h-7 w-7 items-center justify-center rounded-md border border-border transition-colors",
              isPaused ? "bg-amber-500/10 text-amber-500 border-amber-500/20" : "text-muted hover:text-active"
            )}
            title={isPaused ? "Resume autoscroll" : "Pause autoscroll"}
          >
            {isPaused ? <Play className="h-3.5 w-3.5" /> : <Pause className="h-3.5 w-3.5" />}
          </button>
          <button
            onClick={handleDownload}
            className="flex h-7 w-7 items-center justify-center rounded-md border border-border text-muted transition-colors hover:text-active"
            title="Download logs"
          >
            <Download className="h-3.5 w-3.5" />
          </button>
          <div className="h-4 w-px bg-border mx-1" />
          <div className="flex items-center gap-1.5 overflow-x-auto">
            <AnimatePresence>
            {streams.map((name) => (
              <motion.button
                key={name}
                layout
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                onClick={() => onStreamChange(name)}
                className={cn(
                  'shrink-0 rounded-md border px-2.5 py-1 font-mono text-[10px] uppercase transition-colors',
                  current === name
                    ? 'border-accent bg-accent/10 text-accent'
                    : 'border-border bg-canvas text-muted hover:border-border-glow hover:text-active'
                )}
              >
                {name}
              </motion.button>
            ))}
          </AnimatePresence>
          {current && (
            <button
              onClick={() => onStreamChange(null)}
              className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md border border-border text-muted transition-colors hover:text-active"
              title="Clear selection"
            >
              <X className="h-3 w-3" />
            </button>
            )}
          </div>
        </div>
      </div>

      <div
        ref={scrollRef}
        className="flex-1 overflow-auto rounded-lg bg-black px-4 py-3 font-mono text-xs leading-relaxed"
      >
        {lines.length === 0 ? (
          <span className="text-muted/50">
            {current ? `No data for ${current}` : 'Select a log stream'}
          </span>
        ) : (
          lines.map(({ line }, idx) => (
            <div key={idx} className="flex gap-3">
              <span className="shrink-0 text-accent/60">❯</span>
              <span className="text-accent/90">{line}</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
