import type { LanguageNote } from '../../api/client'

export function LanguageNoteCard({ note }: { note: LanguageNote }) {
  return (
    <article className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">
            {(note.note_type ?? 'note').toLowerCase().replace(/_/g, ' ')}
          </p>
          <h3 className="mt-2 font-display text-2xl text-ink">{note.term ?? 'Language note'}</h3>
          <p className="mt-2 text-sm leading-6 text-ink/72">{note.language ?? 'Language not set'}</p>
        </div>
        <p className="text-xs uppercase tracking-[0.18em] text-ink/55">{formatDateTime(note.updated_at)}</p>
      </div>

      <p className="mt-4 whitespace-pre-wrap text-sm leading-6 text-ink/80">{note.content}</p>
    </article>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
