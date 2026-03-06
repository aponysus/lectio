import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { INQUIRY_STATUSES, type Inquiry, type InquiryInput, type InquiryStatus } from '../../api/client'

const inquiryFormSchema = z.object({
  title: z.string().trim().min(1, 'Title is required').max(240, 'Title must be 240 characters or fewer'),
  question: z.string().trim().min(1, 'Question is required').max(4000, 'Question must be 4000 characters or fewer'),
  status: z.enum(INQUIRY_STATUSES),
  why_it_matters: z.string().max(4000, 'Why it matters must be 4000 characters or fewer'),
  current_view: z.string().max(4000, 'Current view must be 4000 characters or fewer'),
  open_tensions: z.string().max(4000, 'Open tensions must be 4000 characters or fewer'),
})

type InquiryFormValues = z.infer<typeof inquiryFormSchema>

type InquiryFormProps = {
  inquiry?: Inquiry | null
  submitLabel: string
  submitting: boolean
  apiError: string | null
  onSubmit: (input: InquiryInput) => Promise<void>
}

export function InquiryForm({ inquiry, submitLabel, submitting, apiError, onSubmit }: InquiryFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InquiryFormValues>({
    resolver: zodResolver(inquiryFormSchema),
    defaultValues: toFormValues(inquiry),
  })

  useEffect(() => {
    reset(toFormValues(inquiry))
  }, [inquiry, reset])

  const submit = async (values: InquiryFormValues) => {
    await onSubmit({
      title: values.title.trim(),
      question: values.question.trim(),
      status: values.status,
      why_it_matters: values.why_it_matters.trim(),
      current_view: values.current_view.trim(),
      open_tensions: values.open_tensions.trim(),
    })
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit(submit)}>
      <section className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
        <div className="grid gap-5 md:grid-cols-2">
          <Field label="Title" error={errors.title?.message}>
            <input
              {...register('title')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          <Field label="Status" error={errors.status?.message}>
            <select
              {...register('status')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              {INQUIRY_STATUSES.map((status) => (
                <option key={status} value={status}>
                  {status.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <Field className="mt-5" label="Question" error={errors.question?.message}>
          <textarea
            {...register('question')}
            rows={5}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          />
        </Field>
      </section>

      <section className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
        <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Workspace context</p>

        <Field className="mt-5" label="Why it matters" error={errors.why_it_matters?.message}>
          <textarea
            {...register('why_it_matters')}
            rows={4}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          />
        </Field>

        <Field className="mt-5" label="Current view" error={errors.current_view?.message}>
          <textarea
            {...register('current_view')}
            rows={4}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          />
        </Field>

        <Field className="mt-5" label="Open tensions" error={errors.open_tensions?.message}>
          <textarea
            {...register('open_tensions')}
            rows={4}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          />
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

function toFormValues(inquiry?: Inquiry | null): InquiryFormValues {
  return {
    title: inquiry?.title ?? '',
    question: inquiry?.question ?? '',
    status: (inquiry?.status as InquiryStatus | undefined) ?? 'ACTIVE',
    why_it_matters: inquiry?.why_it_matters ?? '',
    current_view: inquiry?.current_view ?? '',
    open_tensions: inquiry?.open_tensions ?? '',
  }
}
