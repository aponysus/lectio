export type BrowseViewMode = 'list' | 'cards'

type ViewModeToggleProps = {
  value: BrowseViewMode
  onChange: (value: BrowseViewMode) => void
}

export function ViewModeToggle({ value, onChange }: ViewModeToggleProps) {
  return (
    <div className="inline-flex rounded-2xl border border-black/10 bg-black/[0.03] p-1">
      {(['list', 'cards'] as const).map((mode) => {
        const active = value === mode

        return (
          <button
            key={mode}
            type="button"
            onClick={() => onChange(mode)}
            aria-pressed={active}
            className={`rounded-[1rem] px-3 py-2 text-sm transition ${
              active ? 'bg-white text-ink shadow-sm' : 'text-ink/68 hover:text-ink'
            }`}
          >
            {mode === 'list' ? 'List' : 'Cards'}
          </button>
        )
      })}
    </div>
  )
}
