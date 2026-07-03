import { useEffect, useRef } from 'react';
import type { LogChunk } from '@/lib/types';

interface LogTerminalProps {
  logs: LogChunk[];
}

export function LogTerminal({ logs }: LogTerminalProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const lines = logs
    .flatMap((chunk) =>
      chunk.data
        .split('\n')
        .filter(Boolean)
        .map((line) => ({ line, name: chunk.name, offset: chunk.offset }))
    )
    .slice(-200);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [lines.length]);

  return (
    <div
      ref={scrollRef}
      className="h-full overflow-auto rounded-lg bg-black px-4 py-3 font-mono text-xs leading-relaxed"
    >
      {lines.length === 0 ? (
        <span className="text-muted/50">No log data yet…</span>
      ) : (
        lines.map(({ line, name }, idx) => (
          <div key={idx} className="flex gap-3">
            <span className="shrink-0 text-muted/40">[{name}]</span>
            <span className="text-accent/90">{line}</span>
          </div>
        ))
      )}
    </div>
  );
}
