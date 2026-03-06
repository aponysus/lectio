import { createContext, type ReactNode, useContext, useState } from 'react'

type ToastTone = 'success' | 'info' | 'error'

type ToastInput = {
  message: string
  tone?: ToastTone
}

type ToastRecord = ToastInput & {
  id: string
}

type ToastContextValue = {
  showToast: (input: ToastInput) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastRecord[]>([])

  const dismissToast = (id: string) => {
    setToasts((current) => current.filter((toast) => toast.id !== id))
  }

  const showToast = ({ message, tone = 'success' }: ToastInput) => {
    const id = `${Date.now()}-${Math.random().toString(16).slice(2)}`
    setToasts((current) => [...current, { id, message, tone }])

    window.setTimeout(() => {
      dismissToast(id)
    }, 3600)
  }

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}

      <div className="pointer-events-none fixed right-4 top-4 z-50 flex w-full max-w-sm flex-col gap-3 sm:right-6 sm:top-6">
        {toasts.map((toast) => (
          <ToastCard key={toast.id} toast={toast} onDismiss={() => dismissToast(toast.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }

  return context
}

function ToastCard({
  toast,
  onDismiss,
}: {
  toast: ToastRecord
  onDismiss: () => void
}) {
  const toneClasses = {
    success: 'border-pine/25 bg-white text-ink shadow-card',
    info: 'border-accent/20 bg-white text-ink shadow-card',
    error: 'border-red-200 bg-red-50 text-red-800 shadow-card',
  }

  return (
    <article
      className={`pointer-events-auto rounded-2xl border px-4 py-3 backdrop-blur ${toneClasses[toast.tone ?? 'success']}`}
      role="status"
      aria-live="polite"
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-sm leading-6">{toast.message}</p>
        <button
          type="button"
          onClick={onDismiss}
          className="rounded-lg px-2 py-1 text-xs uppercase tracking-[0.18em] text-ink/55 transition hover:bg-black/[0.05] hover:text-ink"
        >
          Close
        </button>
      </div>
    </article>
  )
}
