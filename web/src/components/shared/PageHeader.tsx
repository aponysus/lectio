import type { ReactNode } from 'react'

type PageHeaderProps = {
  eyebrow: string
  title: string
  description: string
  actions?: ReactNode
}

export function PageHeader({ eyebrow, title, description, actions }: PageHeaderProps) {
  return (
    <section className="rounded-[1.5rem] border border-black/5 bg-white/70 px-5 py-6 shadow-card backdrop-blur lg:px-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-3xl">
          <p className="text-[0.7rem] uppercase tracking-[0.24em] text-accent/80">{eyebrow}</p>
          <h2 className="mt-2 font-display text-[2rem] leading-tight text-ink">{title}</h2>
          <p className="mt-3 text-sm leading-6 text-ink/72">{description}</p>
        </div>
        {actions ? <div className="flex flex-wrap gap-3">{actions}</div> : null}
      </div>
    </section>
  )
}
