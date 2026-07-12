import { Languages } from 'lucide-react';
import { useI18n, type Language } from '@/lib/i18n';
import { cn } from '@/lib/utils';

const options: Array<{ language: Language; label: string }> = [
  { language: 'en', label: 'EN' },
  { language: 'ru', label: 'RU' },
];

export function LanguageSwitcher({ className }: { className?: string }) {
  const { language, setLanguage, t } = useI18n();

  return (
    <div
      className={cn(
        'flex h-9 items-center gap-1 rounded-xl border border-white/10 bg-white/[0.035] p-1',
        className
      )}
      aria-label={t('common.language')}
    >
      <Languages className="ml-1 h-3.5 w-3.5 text-muted-soft" aria-hidden="true" />
      {options.map((option) => {
        const active = language === option.language;
        return (
          <button
            key={option.language}
            type="button"
            onClick={() => setLanguage(option.language)}
            className={cn(
              'h-7 rounded-lg px-2 font-mono text-[10px] font-semibold transition-colors',
              active ? 'bg-white text-black' : 'text-muted-soft hover:bg-white/[0.06] hover:text-active'
            )}
            aria-pressed={active}
            title={option.language === 'en' ? t('common.english') : t('common.russian')}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
