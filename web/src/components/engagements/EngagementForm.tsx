import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useEffect, useState } from 'react'
import { useFieldArray, useForm } from 'react-hook-form'
import { z } from 'zod'
import {
  ACCESS_MODES,
  CLAIM_STATUSES,
  CLAIM_TYPES,
  INQUIRY_STATUSES,
  LANGUAGE_NOTE_TYPES,
  type AccessMode,
  type Claim,
  type ClaimStatus,
  type ClaimType,
  type Engagement,
  type EngagementInput,
  type Inquiry,
  type InquiryInput,
  type LanguageNote,
  type LanguageNoteType,
  type Source,
} from '../../api/client'

const claimDraftSchema = z
  .object({
    claim_id: z.string().optional(),
    text: z.string().max(4000, 'Claim text must be 4000 characters or fewer'),
    claim_type: z.union([z.literal(''), z.enum(CLAIM_TYPES)]),
    confidence: z
      .string()
      .refine((value) => value === '' || /^[1-5]$/.test(value), 'Confidence must be between 1 and 5'),
    status: z.union([z.literal(''), z.enum(CLAIM_STATUSES)]),
    notes: z.string().max(4000, 'Claim notes must be 4000 characters or fewer'),
  })
  .superRefine((value, ctx) => {
    const hasAnyValue =
      value.text.trim() !== '' ||
      value.claim_type !== '' ||
      value.confidence !== '' ||
      value.status !== '' ||
      value.notes.trim() !== ''

    if (!hasAnyValue) {
      return
    }

    if (value.text.trim() === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['text'], message: 'Claim text is required' })
    }
    if (value.claim_type === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['claim_type'], message: 'Claim type is required' })
    }
    if (value.status === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['status'], message: 'Claim status is required' })
    }
  })

const languageNoteDraftSchema = z
  .object({
    note_id: z.string().optional(),
    term: z.string().max(240, 'Term must be 240 characters or fewer'),
    language: z.string().max(64, 'Language must be 64 characters or fewer'),
    note_type: z.union([z.literal(''), z.enum(LANGUAGE_NOTE_TYPES)]),
    content: z.string().max(4000, 'Language note content must be 4000 characters or fewer'),
  })
  .superRefine((value, ctx) => {
    const hasAnyValue =
      value.term.trim() !== '' ||
      value.language.trim() !== '' ||
      value.note_type !== '' ||
      value.content.trim() !== ''

    if (!hasAnyValue) {
      return
    }

    if (value.language.trim() === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['language'], message: 'Language is required' })
    }
    if (value.note_type === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['note_type'], message: 'Note type is required' })
    }
    if (value.content.trim() === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['content'], message: 'Content is required' })
    }
  })

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
  claims: z.array(claimDraftSchema).max(3, 'You can add at most three claims during capture'),
  language_notes: z.array(languageNoteDraftSchema),
})

type EngagementFormValues = z.infer<typeof engagementFormSchema>

export type EngagementClaimSubmission = {
  claim_id?: string
  text: string
  claim_type: ClaimType
  confidence: number | null
  status: ClaimStatus
  notes: string
}

export type EngagementFormLanguageNoteSubmission = {
  note_id?: string
  term: string
  language: string
  note_type: LanguageNoteType
  content: string
}

export type EngagementFormSubmission = {
  engagement: EngagementInput
  inquiry_ids: string[]
  claims: EngagementClaimSubmission[]
  language_notes: EngagementFormLanguageNoteSubmission[]
}

type EngagementFormProps = {
  engagement?: Engagement | null
  claims: Claim[]
  languageNotes: LanguageNote[]
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

const baseFieldClassName =
  'w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent focus:ring-2 focus:ring-accent/20'

const nestedFieldClassName =
  'w-full rounded-2xl border border-black/10 bg-white/85 px-4 py-3 outline-none transition focus:border-accent focus:ring-2 focus:ring-accent/20'

export function EngagementForm({
  engagement,
  claims,
  languageNotes,
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
  const [openSections, setOpenSections] = useState(() => ({
    inquiries: inquiries.length === 0 || linkedInquiryIDs.length > 0,
    claims: claims.length > 0,
    advanced: hasSeededAdvancedData(engagement, languageNotes),
  }))

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    getValues,
    control,
    formState: { errors },
  } = useForm<EngagementFormValues>({
    resolver: zodResolver(engagementFormSchema),
    defaultValues: toFormValues(engagement, claims, languageNotes, defaultSourceID, linkedInquiryIDs),
  })

  useEffect(() => {
    reset(toFormValues(engagement, claims, languageNotes, defaultSourceID, linkedInquiryIDs))
  }, [claims, defaultSourceID, engagement, languageNotes, linkedInquiryIDs, reset])

  const claimFields = useFieldArray({
    control,
    name: 'claims',
  })
  const languageNoteFields = useFieldArray({
    control,
    name: 'language_notes',
  })
  const selectedInquiryIDs = watch('inquiry_ids')
  const draftedClaims = watch('claims')
  const draftedLanguageNotes = watch('language_notes')
  const sourceLanguage = watch('source_language')
  const reflectionLanguage = watch('reflection_language')
  const accessMode = watch('access_mode')
  const revisitPriority = watch('revisit_priority')
  const isRereadOrRewatch = watch('is_reread_or_rewatch')

  useEffect(() => {
    if (errors.claims) {
      setOpenSections((current) => (current.claims ? current : { ...current, claims: true }))
    }
  }, [errors.claims])

  useEffect(() => {
    if (errors.language_notes || errors.source_language || errors.reflection_language || errors.access_mode || errors.revisit_priority) {
      setOpenSections((current) => (current.advanced ? current : { ...current, advanced: true }))
    }
  }, [errors.access_mode, errors.language_notes, errors.reflection_language, errors.revisit_priority, errors.source_language])

  useEffect(() => {
    if (inlineInquiryError) {
      setOpenSections((current) => (current.inquiries ? current : { ...current, inquiries: true }))
    }
  }, [inlineInquiryError])

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
      claims: values.claims
        .filter((claim) => hasAnyClaimValue(claim))
        .map((claim) => ({
          claim_id: claim.claim_id,
          text: claim.text.trim(),
          claim_type: claim.claim_type as ClaimType,
          confidence: claim.confidence === '' ? null : Number(claim.confidence),
          status: claim.status as ClaimStatus,
          notes: claim.notes.trim(),
        })),
      language_notes: values.language_notes
        .filter((note) => hasAnyLanguageNoteValue(note))
        .map((note) => ({
          note_id: note.note_id,
          term: note.term.trim(),
          language: note.language.trim(),
          note_type: note.note_type as LanguageNoteType,
          content: note.content.trim(),
        })),
    })
  }

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

      const nextSelectedInquiryIDs = getValues('inquiry_ids')
      setValue('inquiry_ids', [...nextSelectedInquiryIDs, inquiry.id], { shouldDirty: true })
      setInlineInquiryTitle('')
      setInlineInquiryQuestion('')
      setInlineInquiryWhyItMatters('')
    } catch (err) {
      setInlineInquiryError(err instanceof Error ? err.message : 'Failed to create inquiry')
    } finally {
      setCreatingInquiry(false)
    }
  }

  const claimCount = draftedClaims.filter((claim) => hasAnyClaimValue(claim)).length
  const languageNoteCount = draftedLanguageNotes.filter((note) => hasAnyLanguageNoteValue(note)).length
  const contextFieldCount = [
    sourceLanguage.trim() !== '',
    reflectionLanguage.trim() !== '',
    accessMode !== '',
    revisitPriority !== '',
    isRereadOrRewatch,
  ].filter(Boolean).length

  const inquirySummary =
    selectedInquiryIDs.length === 0
      ? 'No linked inquiries yet. Open this only if the engagement should feed a live question.'
      : `${selectedInquiryIDs.length} ${pluralize(selectedInquiryIDs.length, 'linked inquiry')} selected.`

  const claimSummary =
    claimCount === 0
      ? 'Optional. Add up to three claims only when the reflection has already sharpened.'
      : `${claimCount} ${pluralize(claimCount, 'claim draft')} ready to save.`

  const advancedSummary =
    languageNoteCount === 0 && contextFieldCount === 0
      ? 'Collapsed by default. Open only when language, access mode, or revisit metadata materially matters.'
      : `${[
          languageNoteCount > 0 ? `${languageNoteCount} ${pluralize(languageNoteCount, 'language note')}` : null,
          contextFieldCount > 0 ? `${contextFieldCount} ${pluralize(contextFieldCount, 'context field')}` : null,
        ]
          .filter(Boolean)
          .join(' • ')} configured.`

  const toggleSection = (section: keyof typeof openSections) => {
    setOpenSections((current) => ({ ...current, [section]: !current[section] }))
  }

  return (
    <form className="space-y-5" onSubmit={handleSubmit(submit)}>
      <section className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-3xl">
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Core capture</p>
            <h3 className="mt-2 font-display text-[1.7rem] leading-tight text-ink">Start with the encounter itself</h3>
            <p className="mt-3 text-sm leading-6 text-ink/74">
              Keep the default path short: source, time, reflection, and a brief note on why the encounter mattered.
            </p>
          </div>
          <p className="rounded-full bg-black/[0.04] px-3 py-2 text-xs uppercase tracking-[0.18em] text-ink/62">
            Required first
          </p>
        </div>

        <div className="mt-5 grid gap-5 md:grid-cols-2">
          <Field label="Source" error={errors.source_id?.message}>
            <select {...register('source_id')} className={baseFieldClassName}>
              <option value="">Select a source</option>
              {sources.map((source) => (
                <option key={source.id} value={source.id}>
                  {source.title} • {source.medium.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Engaged at" error={errors.engaged_at?.message}>
            <input {...register('engaged_at')} type="datetime-local" className={baseFieldClassName} />
          </Field>

          <Field label="Portion label" error={errors.portion_label?.message}>
            <input
              {...register('portion_label')}
              placeholder="Chapter 3, pages 80-115, episodes 1-3..."
              className={baseFieldClassName}
            />
          </Field>

          <Field label="Why it matters" error={errors.why_it_matters?.message}>
            <input {...register('why_it_matters')} placeholder="Optional, but encouraged" className={baseFieldClassName} />
          </Field>
        </div>

        <Field className="mt-5" label="Reflection" error={errors.reflection?.message}>
          <textarea {...register('reflection')} rows={10} className={baseFieldClassName} />
        </Field>
      </section>

      <DisclosureSection
        eyebrow="Inquiries"
        title="Attach this engagement to live questions"
        description="Link the engagement to the questions it should feed, or create a new inquiry inline if the work just opened one."
        summary={inquirySummary}
        open={openSections.inquiries}
        onToggle={() => toggleSection('inquiries')}
      >
        {inquiries.length === 0 ? (
          <p className="rounded-2xl bg-black/[0.03] px-4 py-4 text-sm leading-6 text-ink/72">
            No inquiries exist yet. Use the inline form below to create the first one without leaving capture.
          </p>
        ) : (
          <div className="grid gap-3">
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

        <div className="mt-5 rounded-[1.5rem] border border-black/8 bg-canvas/70 p-5">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Inline inquiry create</p>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <Field label="Title">
              <input value={inlineInquiryTitle} onChange={(event) => setInlineInquiryTitle(event.target.value)} className={nestedFieldClassName} />
            </Field>
            <Field label="Why it matters">
              <input
                value={inlineInquiryWhyItMatters}
                onChange={(event) => setInlineInquiryWhyItMatters(event.target.value)}
                className={nestedFieldClassName}
              />
            </Field>
          </div>

          <Field className="mt-4" label="Question">
            <textarea
              value={inlineInquiryQuestion}
              onChange={(event) => setInlineInquiryQuestion(event.target.value)}
              rows={4}
              className={nestedFieldClassName}
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
      </DisclosureSection>

      <DisclosureSection
        eyebrow="Claims"
        title="Extract one to three sharper takeaways"
        description="Keep this optional and lightweight. Claims can be tentative, personal, interpretive, or openly framed as questions."
        summary={claimSummary}
        open={openSections.claims}
        onToggle={() => toggleSection('claims')}
        actions={
          claimFields.fields.length < 3 ? (
            <button
              type="button"
              onClick={() => {
                claimFields.append(emptyClaimDraft())
                setOpenSections((current) => ({ ...current, claims: true }))
              }}
              className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              Add claim
            </button>
          ) : null
        }
      >
        {claimFields.fields.length === 0 ? (
          <p className="rounded-2xl bg-black/[0.03] px-4 py-4 text-sm leading-6 text-ink/72">
            No inline claims yet. Add one only if the reflection has already sharpened into a proposition or question.
          </p>
        ) : (
          <div className="space-y-4">
            {claimFields.fields.map((field, index) => (
              <article key={field.id} className="rounded-[1.5rem] border border-black/8 bg-canvas/60 p-5">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Claim {index + 1}</p>
                    <p className="mt-2 text-sm leading-6 text-ink/72">
                      Keep it concise enough to revise later without rereading the whole reflection.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => claimFields.remove(index)}
                    className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 transition hover:bg-red-100"
                  >
                    Remove
                  </button>
                </div>

                <Field className="mt-4" label="Claim text" error={errors.claims?.[index]?.text?.message}>
                  <textarea {...register(`claims.${index}.text`)} rows={4} className={nestedFieldClassName} />
                </Field>

                <div className="mt-4 grid gap-4 md:grid-cols-3">
                  <Field label="Claim type" error={errors.claims?.[index]?.claim_type?.message}>
                    <select {...register(`claims.${index}.claim_type`)} className={nestedFieldClassName}>
                      <option value="">Select type</option>
                      {CLAIM_TYPES.map((claimType) => (
                        <option key={claimType} value={claimType}>
                          {claimType.toLowerCase().replace(/_/g, ' ')}
                        </option>
                      ))}
                    </select>
                  </Field>

                  <Field label="Status" error={errors.claims?.[index]?.status?.message}>
                    <select {...register(`claims.${index}.status`)} className={nestedFieldClassName}>
                      <option value="">Select status</option>
                      {CLAIM_STATUSES.map((status) => (
                        <option key={status} value={status}>
                          {status.toLowerCase().replace(/_/g, ' ')}
                        </option>
                      ))}
                    </select>
                  </Field>

                  <Field label="Confidence" error={errors.claims?.[index]?.confidence?.message}>
                    <select {...register(`claims.${index}.confidence`)} className={nestedFieldClassName}>
                      <option value="">Not set</option>
                      {[1, 2, 3, 4, 5].map((confidence) => (
                        <option key={confidence} value={confidence}>
                          {confidence}
                        </option>
                      ))}
                    </select>
                  </Field>
                </div>

                <Field className="mt-4" label="Notes" error={errors.claims?.[index]?.notes?.message}>
                  <textarea {...register(`claims.${index}.notes`)} rows={3} className={nestedFieldClassName} />
                </Field>
              </article>
            ))}
          </div>
        )}
      </DisclosureSection>

      <DisclosureSection
        eyebrow="Advanced"
        title="Language and revisit metadata"
        description="Open this only when wording, translation, access mode, or revisit intent materially changes how the engagement should be understood later."
        summary={advancedSummary}
        open={openSections.advanced}
        onToggle={() => toggleSection('advanced')}
        actions={
          <button
            type="button"
            onClick={() => {
              languageNoteFields.append(emptyLanguageNoteDraft())
              setOpenSections((current) => ({ ...current, advanced: true }))
            }}
            className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
          >
            Add language note
          </button>
        }
      >
        <div className="space-y-5">
          <div>
            <p className="text-xs uppercase tracking-[0.22em] text-accent/78">Language notes</p>
            <p className="mt-2 text-sm leading-6 text-ink/72">
              Use this only when language materially changes the encounter. A brief note about a term, register, or
              cultural nuance is enough.
            </p>
          </div>

          {languageNoteFields.fields.length === 0 ? (
            <p className="rounded-2xl bg-black/[0.03] px-4 py-4 text-sm leading-6 text-ink/72">
              No language notes yet. Add one when wording, translation, or cultural semantics shaped the encounter.
            </p>
          ) : (
            <div className="space-y-4">
              {languageNoteFields.fields.map((field, index) => (
                <article key={field.id} className="rounded-[1.5rem] border border-black/8 bg-canvas/60 p-5">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Language note {index + 1}</p>
                      <p className="mt-2 text-sm leading-6 text-ink/72">
                        Keep it concrete enough that the note remains useful without reopening the source.
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => languageNoteFields.remove(index)}
                      className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 transition hover:bg-red-100"
                    >
                      Remove
                    </button>
                  </div>

                  <div className="mt-4 grid gap-4 md:grid-cols-3">
                    <Field label="Term (optional)" error={errors.language_notes?.[index]?.term?.message}>
                      <input {...register(`language_notes.${index}.term`)} className={nestedFieldClassName} />
                    </Field>

                    <Field label="Language" error={errors.language_notes?.[index]?.language?.message}>
                      <input {...register(`language_notes.${index}.language`)} placeholder="zh, ar, fr..." className={nestedFieldClassName} />
                    </Field>

                    <Field label="Note type" error={errors.language_notes?.[index]?.note_type?.message}>
                      <select {...register(`language_notes.${index}.note_type`)} className={nestedFieldClassName}>
                        <option value="">Select type</option>
                        {LANGUAGE_NOTE_TYPES.map((noteType) => (
                          <option key={noteType} value={noteType}>
                            {noteType.toLowerCase().replace(/_/g, ' ')}
                          </option>
                        ))}
                      </select>
                    </Field>
                  </div>

                  <Field className="mt-4" label="Content" error={errors.language_notes?.[index]?.content?.message}>
                    <textarea {...register(`language_notes.${index}.content`)} rows={4} className={nestedFieldClassName} />
                  </Field>
                </article>
              ))}
            </div>
          )}

          <div className="border-t border-black/6 pt-5">
            <p className="text-xs uppercase tracking-[0.22em] text-accent/78">Context</p>
            <div className="mt-4 grid gap-5 md:grid-cols-2">
              <Field label="Source language" error={errors.source_language?.message}>
                <input {...register('source_language')} className={baseFieldClassName} />
              </Field>

              <Field label="Reflection language" error={errors.reflection_language?.message}>
                <input {...register('reflection_language')} className={baseFieldClassName} />
              </Field>

              <Field label="Access mode" error={errors.access_mode?.message}>
                <select {...register('access_mode')} className={baseFieldClassName}>
                  <option value="">Not set</option>
                  {ACCESS_MODES.map((mode) => (
                    <option key={mode} value={mode}>
                      {mode.toLowerCase().replace(/_/g, ' ')}
                    </option>
                  ))}
                </select>
              </Field>

              <Field label="Revisit priority" error={errors.revisit_priority?.message}>
                <select {...register('revisit_priority')} className={baseFieldClassName}>
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
          </div>
        </div>
      </DisclosureSection>

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

function DisclosureSection({
  eyebrow,
  title,
  description,
  summary,
  open,
  onToggle,
  actions,
  children,
}: {
  eyebrow: string
  title: string
  description: string
  summary: string
  open: boolean
  onToggle: () => void
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-3xl">
          <p className="text-xs uppercase tracking-[0.24em] text-accent/80">{eyebrow}</p>
          <h3 className="mt-2 font-display text-[1.7rem] leading-tight text-ink">{title}</h3>
          <p className="mt-3 text-sm leading-6 text-ink/74">{open ? description : summary}</p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {actions}
          <button
            type="button"
            onClick={onToggle}
            aria-expanded={open}
            className="rounded-2xl border border-black/10 bg-black/[0.03] px-4 py-3 text-sm text-ink transition hover:bg-black/[0.05]"
          >
            {open ? 'Collapse' : 'Expand'}
          </button>
        </div>
      </div>

      {open ? <div className="mt-5">{children}</div> : null}
    </section>
  )
}

function toFormValues(
  engagement?: Engagement | null,
  claims: Claim[] = [],
  languageNotes: LanguageNote[] = [],
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
    claims: claims.map((claim) => ({
      claim_id: claim.id,
      text: claim.text,
      claim_type: claim.claim_type,
      confidence: claim.confidence ? String(claim.confidence) : '',
      status: claim.status,
      notes: claim.notes ?? '',
    })),
    language_notes: languageNotes.map((note) => ({
      note_id: note.id,
      term: note.term ?? '',
      language: note.language ?? '',
      note_type: note.note_type ?? '',
      content: note.content,
    })),
  }
}

function emptyClaimDraft(): EngagementFormValues['claims'][number] {
  return {
    claim_id: undefined,
    text: '',
    claim_type: '',
    confidence: '',
    status: '',
    notes: '',
  }
}

function hasAnyClaimValue(claim: EngagementFormValues['claims'][number]) {
  return (
    claim.text.trim() !== '' ||
    claim.claim_type !== '' ||
    claim.confidence !== '' ||
    claim.status !== '' ||
    claim.notes.trim() !== ''
  )
}

function emptyLanguageNoteDraft(): EngagementFormValues['language_notes'][number] {
  return {
    note_id: undefined,
    term: '',
    language: '',
    note_type: '',
    content: '',
  }
}

function hasAnyLanguageNoteValue(note: EngagementFormValues['language_notes'][number]) {
  return (
    note.term.trim() !== '' ||
    note.language.trim() !== '' ||
    note.note_type !== '' ||
    note.content.trim() !== ''
  )
}

function hasSeededAdvancedData(engagement?: Engagement | null, languageNotes: LanguageNote[] = []) {
  return Boolean(
    languageNotes.length > 0 ||
      engagement?.source_language ||
      engagement?.reflection_language ||
      engagement?.access_mode ||
      engagement?.revisit_priority ||
      engagement?.is_reread_or_rewatch,
  )
}

function pluralize(count: number, singular: string, plural = `${singular}s`) {
  return count === 1 ? singular : plural
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
