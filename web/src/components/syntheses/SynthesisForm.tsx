import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { SYNTHESIS_TYPES, type Inquiry, type Synthesis, type SynthesisInput, type SynthesisType } from '../../api/client'
import { formFieldClassName } from '../shared/formStyles'

const synthesisFormSchema = z.object({
  title: z.string().trim().min(1, 'Title is required').max(240, 'Title must be 240 characters or fewer'),
  type: z.enum(SYNTHESIS_TYPES),
  body: z.string().trim().min(1, 'Body is required').max(20000, 'Body must be 20000 characters or fewer'),
  notes: z.string().max(4000, 'Notes must be 4000 characters or fewer'),
})

type SynthesisFormValues = z.infer<typeof synthesisFormSchema>

type SynthesisFormProps = {
  synthesis?: Synthesis | null
  inquiry: Inquiry
  submitLabel: string
  submitting: boolean
  apiError: string | null
  onSubmit: (input: SynthesisInput) => Promise<void>
}

export function SynthesisForm({ synthesis, inquiry, submitLabel, submitting, apiError, onSubmit }: SynthesisFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SynthesisFormValues>({
    resolver: zodResolver(synthesisFormSchema),
    defaultValues: toFormValues(synthesis, inquiry),
  })

  useEffect(() => {
    reset(toFormValues(synthesis, inquiry))
  }, [inquiry, reset, synthesis])

  const submit = async (values: SynthesisFormValues) => {
    await onSubmit({
      title: values.title.trim(),
      type: values.type,
      body: values.body.trim(),
      inquiry_id: inquiry.id,
      notes: values.notes.trim(),
    })
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit(submit)}>
      <section className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-3xl">
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Compression pass</p>
            <h3 className="mt-2 font-display text-[1.7rem] leading-tight text-ink">Write the smallest synthesis that changes the inquiry</h3>
            <p className="mt-3 text-sm leading-6 text-ink/74">
              Aim for a working position, not a definitive conclusion. It should clarify what you think and what remains unresolved.
            </p>
          </div>
        </div>

        <div className="grid gap-5 md:grid-cols-2">
          <Field label="Title" error={errors.title?.message}>
            <input {...register('title')} className={formFieldClassName} />
          </Field>

          <Field label="Type" error={errors.type?.message}>
            <select {...register('type')} className={formFieldClassName}>
              {SYNTHESIS_TYPES.map((type) => (
                <option key={type} value={type}>
                  {type.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <Field className="mt-5" label="Body" error={errors.body?.message}>
          <textarea {...register('body')} rows={14} className={formFieldClassName} />
        </Field>
      </section>

      <section className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
        <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Linked inquiry</p>
        <h3 className="mt-2 font-display text-[1.7rem] leading-tight text-ink">{inquiry.title}</h3>
        <p className="mt-3 text-sm leading-6 text-ink/78">{inquiry.question}</p>

        {inquiry.why_it_matters ? (
          <div className="mt-5 rounded-2xl bg-black/[0.03] px-4 py-4">
            <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Why it matters</p>
            <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{inquiry.why_it_matters}</p>
          </div>
        ) : null}

        <Field className="mt-5" label="Notes" error={errors.notes?.message}>
          <textarea {...register('notes')} rows={5} className={formFieldClassName} />
        </Field>
      </section>

      {apiError ? (
        <p className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">{apiError}</p>
      ) : null}

      <div className="flex justify-end">
        <button
          type="submit"
          disabled={submitting}
          className="rounded-2xl bg-pine px-5 py-3 text-sm font-medium text-white transition hover:bg-pine/90 disabled:cursor-wait disabled:opacity-70"
        >
          {submitting ? 'Saving...' : submitLabel}
        </button>
      </div>
    </form>
  )
}

function Field({
  label,
  error,
  className,
  children,
}: {
  label: string
  error?: string
  className?: string
  children: ReactNode
}) {
  return (
    <label className={`block ${className ?? ''}`}>
      <span className="mb-2 block text-sm text-ink/75">{label}</span>
      {children}
      {error ? <span className="mt-2 block text-sm text-amber-700">{error}</span> : null}
    </label>
  )
}

function toFormValues(synthesis: Synthesis | null | undefined, inquiry: Inquiry): SynthesisFormValues {
  return {
    title: synthesis?.title ?? `${inquiry.title}: checkpoint`,
    type: (synthesis?.type as SynthesisType | undefined) ?? 'CHECKPOINT',
    body: synthesis?.body ?? '',
    notes: synthesis?.notes ?? '',
  }
}
