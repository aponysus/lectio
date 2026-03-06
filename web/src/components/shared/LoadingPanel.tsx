type LoadingPanelProps = {
  label: string
  variant?: 'page' | 'list' | 'shell'
}

export function LoadingPanel({ label, variant = 'page' }: LoadingPanelProps) {
  if (variant === 'shell') {
    return (
      <div
        className="w-full max-w-sm rounded-[1.5rem] border border-black/5 bg-white/80 px-6 py-5 shadow-card"
        role="status"
        aria-live="polite"
      >
        <span className="sr-only">{label}</span>
        <div className="animate-pulse space-y-3">
          <div className="h-3 w-20 rounded-full bg-black/[0.08]" />
          <div className="h-6 w-40 rounded-full bg-black/[0.12]" />
          <div className="h-4 w-32 rounded-full bg-black/[0.08]" />
        </div>
      </div>
    )
  }

  if (variant === 'list') {
    return (
      <section
        className="rounded-[1.5rem] border border-black/5 bg-white/70 px-5 py-5 shadow-card backdrop-blur"
        role="status"
        aria-live="polite"
      >
        <span className="sr-only">{label}</span>
        <div className="animate-pulse space-y-5">
          <div className="h-4 w-28 rounded-full bg-black/[0.08]" />
          <div className="grid gap-4 xl:grid-cols-2">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={index} className="rounded-[1.25rem] bg-black/[0.035] px-4 py-4">
                <div className="h-3 w-24 rounded-full bg-black/[0.08]" />
                <div className="mt-3 h-6 w-3/4 rounded-full bg-black/[0.1]" />
                <div className="mt-4 space-y-2">
                  <div className="h-4 w-full rounded-full bg-black/[0.06]" />
                  <div className="h-4 w-5/6 rounded-full bg-black/[0.06]" />
                  <div className="h-4 w-2/3 rounded-full bg-black/[0.06]" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>
    )
  }

  return (
    <section
      className="rounded-[1.5rem] border border-black/5 bg-white/70 px-5 py-5 shadow-card backdrop-blur"
      role="status"
      aria-live="polite"
    >
      <span className="sr-only">{label}</span>
      <div className="animate-pulse space-y-5">
        <div className="space-y-3">
          <div className="h-3 w-24 rounded-full bg-black/[0.08]" />
          <div className="h-8 w-2/3 rounded-full bg-black/[0.12]" />
          <div className="h-4 w-full max-w-2xl rounded-full bg-black/[0.07]" />
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          <div className="rounded-[1.25rem] bg-black/[0.035] px-4 py-4">
            <div className="h-3 w-24 rounded-full bg-black/[0.08]" />
            <div className="mt-4 space-y-3">
              <div className="h-4 w-full rounded-full bg-black/[0.06]" />
              <div className="h-4 w-11/12 rounded-full bg-black/[0.06]" />
              <div className="h-4 w-4/5 rounded-full bg-black/[0.06]" />
              <div className="h-20 w-full rounded-[1rem] bg-black/[0.06]" />
            </div>
          </div>

          <div className="rounded-[1.25rem] bg-black/[0.035] px-4 py-4">
            <div className="h-3 w-20 rounded-full bg-black/[0.08]" />
            <div className="mt-4 space-y-3">
              <div className="h-10 w-full rounded-[1rem] bg-black/[0.06]" />
              <div className="h-10 w-full rounded-[1rem] bg-black/[0.06]" />
              <div className="h-10 w-full rounded-[1rem] bg-black/[0.06]" />
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
