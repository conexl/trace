import * as React from 'react';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { MoreHorizontal, ScrollText, Settings, RotateCw, Square, Trash2 } from 'lucide-react';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { cn } from '@/lib/utils';

interface ServiceActionMenuProps {
  serviceName?: string;
  onViewLogs: () => void;
  onEditPolicy: () => void;
  onRestart: () => void;
  onStop: () => void;
  onRemove: () => void;
  canControl?: boolean;
}

export function ServiceActionMenu({
  serviceName,
  onViewLogs,
  onEditPolicy,
  onRestart,
  onStop,
  onRemove,
  canControl = false,
}: ServiceActionMenuProps) {
  const [confirmRemove, setConfirmRemove] = React.useState(false);

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded-md border border-border bg-canvas',
            'text-muted transition-all duration-200',
            'hover:border-accent hover:text-accent hover:bg-accent/5',
            'focus:outline-none'
          )}
        >
          <MoreHorizontal className="h-4 w-4" />
        </button>
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          sideOffset={6}
          align="end"
          className={cn(
            'z-50 min-w-[180px] rounded-lg border border-border bg-surface p-1 shadow-xl',
            'data-[side=bottom]:animate-in data-[side=top]:animate-in data-[state=open]:fade-in-0'
          )}
        >
          <MenuItem icon={ScrollText} label="View logs" onClick={onViewLogs} />
          <MenuItem icon={Settings} label="Edit policy" onClick={onEditPolicy} />
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <MenuItem icon={RotateCw} label="Restart service" onClick={onRestart} disabled={!canControl} />
          <MenuItem icon={Square} label="Stop service" onClick={onStop} disabled={!canControl} />
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <MenuItem
            icon={Trash2}
            label="Remove from watchdog"
            onClick={() => setConfirmRemove(true)}
            className="text-red-400 hover:bg-red-950/20 hover:text-red-300"
          />
          <DropdownMenu.Arrow className="fill-border" />
        </DropdownMenu.Content>
      </DropdownMenu.Portal>

      <ConfirmationDialog
        open={confirmRemove}
        onOpenChange={setConfirmRemove}
        title="Remove from watchdog"
        description={
          serviceName
            ? `Stop monitoring "${serviceName}"? You can add it back later.`
            : 'Stop monitoring this service? You can add it back later.'
        }
        confirmLabel="Remove"
        variant="danger"
        onConfirm={onRemove}
      />
    </DropdownMenu.Root>
  );
}

function MenuItem({
  icon: Icon,
  label,
  onClick,
  className,
  disabled,
}: {
  icon: React.ElementType;
  label: string;
  onClick: () => void;
  className?: string;
  disabled?: boolean;
}) {
  return (
    <DropdownMenu.Item
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'flex cursor-pointer items-center gap-2 rounded-md px-2.5 py-1.5 text-xs text-active',
        'transition-colors outline-none hover:bg-surface-elevated hover:text-accent',
        'data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40',
        className
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      <span>{label}</span>
    </DropdownMenu.Item>
  );
}
