import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { History, X } from 'lucide-react';

interface Props {
  history: string[];
  current: string;
  onSelect: (term: string) => void;
  onRemove: (term: string) => void;
  onClear: () => void;
}

export function SearchHistoryChips({ history, current, onSelect, onRemove, onClear }: Props) {
  if (history.length === 0) return null;
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      <History className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
      {history.map((term) => {
        const active = term.toLowerCase() === current.trim().toLowerCase();
        return (
          <Badge
            key={term}
            variant={active ? 'default' : 'outline'}
            className="cursor-pointer hover:opacity-80 pr-1 gap-1 group"
            onClick={() => onSelect(term)}
          >
            <span className="truncate max-w-[140px]">{term}</span>
            <button
              type="button"
              aria-label={`remove ${term}`}
              className="ml-0.5 opacity-60 hover:opacity-100"
              onClick={(e) => {
                e.stopPropagation();
                onRemove(term);
              }}
            >
              <X className="w-3 h-3" />
            </button>
          </Badge>
        );
      })}
      <Button
        variant="ghost"
        size="sm"
        className="h-6 px-2 text-xs text-muted-foreground"
        onClick={onClear}
      >
        清空
      </Button>
    </div>
  );
}
