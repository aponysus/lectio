import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import {
  ACCESS_MODES,
  INQUIRY_STATUSES,
  type AccessMode,
  type Engagement,
  type EngagementInput,
  type Inquiry,
  type InquiryInput,
  type Source,
} from '../../api/client'

const engagementFormSchema = z.object({
  source_id: z.string().min(1, 'Source is required'),
  engaged_at: z.string().min(1, 'Engaged at is required'),
  portion_label: z.string().max(240, 'Portion label must be 240 characters or fewer'),
  reflection: z.string().trim().min(1, 'Reflection is required').max(10000, 'Reflection must be 10000 characters or fewer'),
  why_it_matters: z.string().max(4000, 'Why it matters must be 4000 characters or fewer'),
  source_language: z.string().max(64, 'Source language must be 64 characters or fewer'),
  reflection_language: z.string().max(64, 'Reflection language must be 64 characters or fewer'),
  access_mode: z.union([z.literal(''), z.enum(ACCESS_MODES)]),
  revisit_priority: z
    .string()
    .refine((value) => value === '' || /^[1-5]$/.test(value), 'Revisit priority must be between 1 and 5'),
  is_reread_or_rewatch: z.boolean(),
  inquiry_ids: z.array(z.string()),
})

type EngagementFormValues = z.infer<typeof engagementFormSchema>

export type EngagementFormSubmission = {
  engagement: EngagementInput
  inquiry_ids: string[]
}

type EngagementFormProps = {
  engagement?: Engagement | null
  sources: Source[]
  inquiries: Inquiry[]
  linkedInquiryIDs: string[]
  defaultSourceID?: string
  submitLabel: string
  submitting: boolean
  apiError: string | null
  onCreateInquiry: (input: InquiryInput) => Promise<Inquiry>
  onSubmit: (input: EngagementFormSubmission) => Promise<void>
}

export function EngagementForm({
  engagement,
  sources,
  inquiries,
  linkedInquiryIDs,
  defaultSourceID,
  submitLabel,
  submitting,
  apiError,
  onCreateInquiry,
  onSubmit,
}: EngagementFormProps) {
  const [inlineInquiryTitle, setInlineInquiryTitle] = useState('')
  const [inlineInquiryQuestion, setInlineInquiryQuestion] = useState('')
  const [inlineInquiryWhyItMatters, setInlineInquiryWhyItMatters] = useState('')
  const [inlineInquiryError, setInlineInquiryError] = useState<string | null>(null)
  const [creatingInquiry, setCreatingInquiry] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<EngagementFormValues>({
    resolver: zodResolver(engagementFormSchema),
    defaultValues: toFormValues(engagement, defaultSourceID, linkedInquiryIDs),
  })

  useEffect(() => {
    reset(toFormValues(engagement, defaultSourceID, linkedInquiryIDs))
  }, [defaultSourceID, engagement, linkedInquiryIDs, reset])

  const submit = async (values: EngagementFormValues) => {
    await onSubmit({
      engagement: {
        source_id: values.source_id,
        engaged_at: new Date(values.engaged_at).toISOString(),
        portion_label: values.portion_label.trim(),
        reflection: values.reflection.trim(),
        why_it_matters: values.why_it_matters.trim(),
        source_language: values.source_language.trim(),
        reflection_language: values.reflection_language.trim(),
        access_mode: values.access_mode,
        revisit_priority: values.revisit_priority === '' ? null : Number(values.revisit_priority),
        is_reread_or_rewatch: values.is_reread_or_rewatch,
      },
      inquiry_ids: values.inquiry_ids,
    })
  }

  const selectedInquiryIDs = watch('inquiry_ids')

  const toggleInquiry = (inquiryID: string) => {
    const nextValue = selectedInquiryIDs.includes(inquiryID)
      ? selectedInquiryIDs.filter((id) => id !== inquiryID)
      : [...selectedInquiryIDs, inquiryID]

    setValue('inquiry_ids', nextValue, { shouldDirty: true })
  }

  const handleInlineInquiryCreate = async () => {
    const title = inlineInquiryTitle.trim()
    const question = inlineInquiryQuestion.trim()
    const whyItMatters = inlineInquiryWhyItMatters.trim()

    if (!title || !question) {
      setInlineInquiryError('Title and question are required for inline inquiry creation.')
      return
    }

    setCreatingInquiry(true)
    setInlineInquiryError(null)

    try {
      const inquiry = await onCreateInquiry({
        title,
        question,
        status: INQUIRY_STATUSES[0],
        why_it_matters: whyItMatters,
        current_view: '',
        open_tensions: '',
      })

      setValue('inquiry_ids', [...selectedInquiryIDs, inquiry.id], { shouldDirty: true })
      setInlineInquiryTitle('')
      setInlineInquiryQuestion('')
      setInlineInquiryWhyItMatters('')
    } catch (err) {
      setInlineInquiryError(err instanceof Error ? err.message : 'Failed to create inquiry')
    } finally {
      setCreatingInquiry(false)
    }
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit(submit)}>
      <section className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
        <div className="grid gap-5 md:grid-cols-2">
          <Field label="Source" error={errors.source_id?.message}>
            <select
              {...register('source_id')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="">Select a source</option>
              {sources.map((source) => (
                <option key={source.id} value={source.id}>
                  {source.title} • {source.medium.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Engaged at" error={errors.engaged_at?.message}>
            <input
              {...register('engaged_at')}
              type="datetime-local"
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          <Field label="Portion label" error={errors.portion_label?.message}>
            <input
              {...register('portion_label')}
              placeholder="Chapter 3, pages 80-115, episodes 1-3..."
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          <Field label="Why it matters" error={errors.why_it_matters?.message}>
            <input
              {...register('why_it_matters')}
              placeholder="Optional, but encouraged"
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>
        </div>

        <Field className="mt-5" label="Reflection" error={errors.reflection?.message}>
          <textarea
            {...register('reflection')}
            rows={10}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          />
        </Field>
      </section>

      <section className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Inquiries</p>
            <h3 className="mt-2 font-display text-3xl text-ink">Attach this engagement to live questions</h3>
            <p className="mt-3 max-w-3xl text-sm leading-6 text-ink/74">
              Keep this lightweight. Link the engagement to the questions it should feed, or create a new inquiry inline
              if the work just opened one.
            </p>
          </div>
        </div>

        {inquiries.length === 0 ? (
          <p className="mt-5 rounded-2xl bg-black/[0.03] px-4 py-4 text-sm leading-6 text-ink/72">
            No inquiries exist yet. Use the inline form below to create the first one without leaving capture.
          </p>
        ) : (
          <div className="mt-5 grid gap-3">
            {inquiries.map((inquiry) => {
              const checked = selectedInquiryIDs.includes(inquiry.id)

              return (
                <label
                  key={inquiry.id}
                  className={`block rounded-2xl border px-4 py-4 transition ${
                    checked ? 'border-pine/35 bg-pine/10' : 'border-black/8 bg-black/[0.03] hover:bg-black/[0.05]'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleInquiry(inquiry.id)}
                      className="mt-1 h-4 w-4 rounded border-black/20"
                    />
                    <div className="min-w-0">
                      <p className="text-sm font-medium text-ink">{inquiry.title}</p>
                      <p className="mt-1 line-clamp-2 text-sm leading-6 text-ink/72">{inquiry.question}</p>
                      <p className="mt-2 text-xs uppercase tracking-[0.2em] text-accent/75">
                        {inquiry.status.toLowerCase().replace(/_/g, ' ')}
                      </p>
                    </div>
                  </div>
                </label>
              )
            })}
          </div>
        )}

        <div className="mt-5 rounded-[1.75rem] border border-black/8 bg-canvas/70 p-5">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Inline inquiry create</p>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <Field label="Title">
              <input
                value={inlineInquiryTitle}
                onChange={(event) => setInlineInquiryTitle(event.target.value)}
                className="w-full rounded-2xl border border-black/10 bg-white/85 px-4 py-3 outline-none transition focus:border-accent"
              />
            </Field>
            <Field label="Why it matters">
              <input
                value={inlineInquiryWhyItMatters}
                onChange={(event) => setInlineInquiryWhyItMatters(event.target.value)}
                className="w-full rounded-2xl border border-black/10 bg-white/85 px-4 py-3 outline-none transition focus:border-accent"
              />
            </Field>
          </div>

          <Field className="mt-4" label="Question">
            <textarea
              value={inlineInquiryQuestion}
              onChange={(event) => setInlineInquiryQuestion(event.target.value)}
              rows={4}
              className="w-full rounded-2xl border border-black/10 bg-white/85 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          {inlineInquiryError ? (
            <p className="mt-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
              {inlineInquiryError}
            </p>
          ) : null}

          <div className="mt-4 flex justify-end">
            <button
              type="button"
              disabled={creatingInquiry}
              onClick={() => void handleInlineInquiryCreate()}
              className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white disabled:cursor-wait disabled:opacity-70"
            >
              {creatingInquiry ? 'Creating inquiry...' : 'Create and attach inquiry'}
            </button>
          </div>
        </div>
      </section>

      <section className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
        <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Context</p>
        <div className="mt-5 grid gap-5 md:grid-cols-2">
          <Field label="Source language" error={errors.source_language?.message}>
            <input
              {...register('source_language')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          <Field label="Reflection language" error={errors.reflection_language?.message}>
            <input
              {...register('reflection_language')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </Field>

          <Field label="Access mode" error={errors.access_mode?.message}>
            <select
              {...register('access_mode')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="">Not set</option>
              {ACCESS_MODES.map((mode) => (
                <option key={mode} value={mode}>
                  {mode.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Revisit priority" error={errors.revisit_priority?.message}>
            <select
              {...register('revisit_priority')}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="">Not set</option>
              {[1, 2, 3, 4, 5].map((priority) => (
                <option key={priority} value={priority}>
                  {priority}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <label className="mt-5 flex items-center gap-3 rounded-2xl bg-black/[0.03] px-4 py-3 text-sm text-ink/80">
          <input type="checkbox" {...register('is_reread_or_rewatch')} className="h-4 w-4 rounded border-black/20" />
          This was a reread, rewatch, or revisit.
        </label>
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

function toFormValues(
  engagement?: Engagement | null,
  defaultSourceID?: string,
  linkedInquiryIDs: string[] = [],
): EngagementFormValues {
  return {
    source_id: engagement?.source_id ?? defaultSourceID ?? '',
    engaged_at: engagement ? toDatetimeLocal(engagement.engaged_at) : toDatetimeLocal(new Date().toISOString()),
    portion_label: engagement?.portion_label ?? '',
    reflection: engagement?.reflection ?? '',
    why_it_matters: engagement?.why_it_matters ?? '',
    source_language: engagement?.source_language ?? '',
    reflection_language: engagement?.reflection_language ?? '',
    access_mode: (engagement?.access_mode as AccessMode | undefined) ?? '',
    revisit_priority: engagement?.revisit_priority ? String(engagement.revisit_priority) : '',
    is_reread_or_rewatch: engagement?.is_reread_or_rewatch ?? false,
    inquiry_ids: linkedInquiryIDs,
  }
}

function toDatetimeLocal(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(
    date.getMinutes(),
  )}`
}
