import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import type { Source, SourceInput, SourceMedium } from '../../api/client'
import { SOURCE_MEDIA } from '../../api/client'
import { formFieldClassName } from '../shared/formStyles'

const sourceFormSchema = z.object({
  title: z.string().trim().min(1, 'Title is required').max(240, 'Title must be 240 characters or fewer'),
  medium: z.enum(SOURCE_MEDIA),
  creator: z.string().max(240, 'Creator must be 240 characters or fewer'),
  year: z
    .string()
    .refine((value) => value === '' || /^-?\d+$/.test(value), 'Year must be an integer')
    .refine(
      (value) => value === '' || (Number(value) >= -4000 && Number(value) <= new Date().getUTCFullYear() + 5),
      'Year must be a reasonable integer',
    ),
  original_language: z.string().max(64, 'Language must be 64 characters or fewer'),
  culture_or_context: z.string().max(160, 'Culture/context must be 160 characters or fewer'),
  notes: z.string().max(4000, 'Notes must be 4000 characters or fewer'),
})

type SourceFormValues = z.infer<typeof sourceFormSchema>

type SourceFormProps = {
  source?: Source | null
  submitLabel: string
  submitting: boolean
  apiError: string | null
  onSubmit: (input: SourceInput) => Promise<void>
}

export function SourceForm({ source, submitLabel, submitting, apiError, onSubmit }: SourceFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SourceFormValues>({
    resolver: zodResolver(sourceFormSchema),
    defaultValues: toFormValues(source),
  })

  useEffect(() => {
    reset(toFormValues(source))
  }, [reset, source])

  const submit = async (values: SourceFormValues) => {
    await onSubmit({
      title: values.title.trim(),
      medium: values.medium,
      creator: values.creator.trim(),
      year: values.year === '' ? null : Number(values.year),
      original_language: values.original_language.trim(),
      culture_or_context: values.culture_or_context.trim(),
      notes: values.notes.trim(),
    })
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit(submit)}>
      <section className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-3xl">
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Source record</p>
            <h3 className="mt-2 font-display text-[1.7rem] leading-tight text-ink">Keep the record clean enough to reuse later</h3>
            <p className="mt-3 text-sm leading-6 text-ink/74">
              Capture only the metadata that will matter again when engagement work starts stacking up.
            </p>
          </div>
        </div>

        <div className="grid gap-5 md:grid-cols-2">
          <Field label="Title" error={errors.title?.message}>
            <input {...register('title')} className={formFieldClassName} />
          </Field>

          <Field label="Medium" error={errors.medium?.message}>
            <select {...register('medium')} className={formFieldClassName}>
              {SOURCE_MEDIA.map((medium) => (
                <option key={medium} value={medium}>
                  {medium.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Creator" error={errors.creator?.message}>
            <input {...register('creator')} className={formFieldClassName} />
          </Field>

          <Field label="Year" error={errors.year?.message}>
            <input {...register('year')} inputMode="numeric" className={formFieldClassName} />
          </Field>

          <Field label="Original language" error={errors.original_language?.message}>
            <input {...register('original_language')} className={formFieldClassName} />
          </Field>

          <Field label="Culture / context" error={errors.culture_or_context?.message}>
            <input {...register('culture_or_context')} className={formFieldClassName} />
          </Field>
        </div>

        <Field className="mt-5" label="Notes" error={errors.notes?.message}>
          <textarea {...register('notes')} rows={6} className={formFieldClassName} />
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

function toFormValues(source?: Source | null): SourceFormValues {
  return {
    title: source?.title ?? '',
    medium: (source?.medium as SourceMedium | undefined) ?? 'BOOK',
    creator: source?.creator ?? '',
    year: source?.year ? String(source.year) : '',
    original_language: source?.original_language ?? '',
    culture_or_context: source?.culture_or_context ?? '',
    notes: source?.notes ?? '',
  }
}
