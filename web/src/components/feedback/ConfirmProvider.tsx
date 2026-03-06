import { createContext, type ReactNode, useContext, useEffect, useState } from 'react'

type ConfirmOptions = {
  title: string
  body: string
  confirmLabel?: string
  cancelLabel?: string
  tone?: 'danger' | 'neutral'
}

type ConfirmDialogState = ConfirmOptions & {
  resolve: (value: boolean) => void
}

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>
}

const ConfirmContext = createContext<ConfirmContextValue | null>(null)

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [dialog, setDialog] = useState<ConfirmDialogState | null>(null)

  useEffect(() => {
    if (!dialog) {
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        closeDialog(false)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [dialog])

  const confirm = (options: ConfirmOptions) =>
    new Promise<boolean>((resolve) => {
      setDialog({
        confirmLabel: 'Confirm',
        cancelLabel: 'Cancel',
        tone: 'danger',
        ...options,
        resolve,
      })
    })

  const closeDialog = (value: boolean) => {
    setDialog((current) => {
      current?.resolve(value)
      return null
    })
  }

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}

      {dialog ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 px-4 py-6 backdrop-blur-sm"
          onClick={() => closeDialog(false)}
        >
          <article
            className="w-full max-w-md rounded-[1.75rem] border border-black/8 bg-white px-6 py-6 shadow-card"
            onClick={(event) => event.stopPropagation()}
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="confirm-dialog-title"
          >
            <p className="text-xs uppercase tracking-[0.22em] text-accent/75">Confirm action</p>
            <h2 id="confirm-dialog-title" className="mt-3 font-display text-3xl text-ink">
              {dialog.title}
            </h2>
            <p className="mt-4 text-sm leading-7 text-ink/74">{dialog.body}</p>

            <div className="mt-6 flex flex-wrap justify-end gap-3">
              <button
                type="button"
                onClick={() => closeDialog(false)}
                className="rounded-2xl border border-black/10 bg-white px-4 py-3 text-sm text-ink transition hover:bg-black/[0.03]"
              >
                {dialog.cancelLabel}
              </button>
              <button
                type="button"
                onClick={() => closeDialog(true)}
                className={`rounded-2xl px-4 py-3 text-sm font-medium text-white transition ${
                  dialog.tone === 'neutral' ? 'bg-ink hover:bg-ink/90' : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {dialog.confirmLabel}
              </button>
            </div>
          </article>
        </div>
      ) : null}
    </ConfirmContext.Provider>
  )
}

export function useConfirm() {
  const context = useContext(ConfirmContext)
  if (!context) {
    throw new Error('useConfirm must be used within a ConfirmProvider')
  }

  return context
}
