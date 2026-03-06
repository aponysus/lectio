export function PlaceholderPage({ title, body }: { title: string; body: string }) {
  return (
    <section className="rounded-[2rem] border border-dashed border-black/10 bg-white/65 px-6 py-8 shadow-card backdrop-blur">
      <p className="text-xs uppercase tracking-[0.3em] text-accent/80">Scaffolded next surface</p>
      <h2 className="mt-3 font-display text-4xl text-ink">{title}</h2>
      <p className="mt-4 max-w-2xl text-base leading-7 text-ink/72">{body}</p>
    </section>
  )
}
