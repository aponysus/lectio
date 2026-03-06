import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  body: string
  action?: ReactNode
}

export function EmptyState({ title, body, action }: EmptyStateProps) {
  return (
    <section className="rounded-[2rem] border border-dashed border-black/10 bg-white/55 px-6 py-8 text-center shadow-card backdrop-blur">
      <h3 className="font-display text-3xl text-ink">{title}</h3>
      <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-ink/70">{body}</p>
      {action ? <div className="mt-6 flex justify-center">{action}</div> : null}
    </section>
  )
}
