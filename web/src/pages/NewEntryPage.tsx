import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { startTransition, useDeferredValue, useState } from 'react'
import { useForm } from 'react-hook-form'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { Link } from 'react-router-dom'
import { apiFetch, isApiError } from '../api/client'

type Entry = {
  id: number
  source_id: number
  passage: string
  reflection: string
  mood: string
  energy: number | null
  tags: Array<{
    id: number
    slug: string
    label: string
  }>
  created_at: string
  updated_at: string
}

type EntryListResponse = {
  data: Entry[]
  meta: {
    page: number
    page_size: number
    total: number
    has_next: boolean
  }
}

type CreateEntryResponse = {
  data: Entry
}

type EntryForm = {
  sourceTitle: string
  sourceAuthor: string
  sourceTradition: string
  passage: string
  reflection: string
  mood: string
  energy: '' | '1' | '2' | '3' | '4' | '5'
  tags: string
}

const defaultValues: EntryForm = {
  sourceTitle: '',
  sourceAuthor: '',
  sourceTradition: '',
  passage: '',
  reflection: '',
  mood: '',
  energy: '',
  tags: '',
}

export function NewEntryPage() {
  const queryClient = useQueryClient()
  const [showDetails, setShowDetails] = useState(false)
  const [savedEntry, setSavedEntry] = useState<Entry | null>(null)
  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<EntryForm>({
    defaultValues,
  })

  const recentEntriesQuery = useQuery({
    queryKey: ['entries', 'new-entry-sidebar'],
    queryFn: () => apiFetch<EntryListResponse>('/entries?page_size=4'),
    retry: false,
  })

  const createMutation = useMutation({
    mutationFn: (values: EntryForm) =>
      apiFetch<CreateEntryResponse>('/entries', {
        method: 'POST',
        body: JSON.stringify({
          source: {
            title: values.sourceTitle.trim(),
            author: values.sourceAuthor.trim(),
            tradition: values.sourceTradition.trim(),
          },
          passage: values.passage.trim(),
          reflection: values.reflection.trim(),
          mood: values.mood.trim(),
          energy: values.energy ? Number(values.energy) : undefined,
          tags: splitTags(values.tags),
        }),
      }),
    onSuccess: async (payload) => {
      await queryClient.invalidateQueries({ queryKey: ['entries'] })
      setSavedEntry(payload.data)
      startTransition(() => {
        reset({
          ...defaultValues,
          sourceTitle: '',
          sourceAuthor: '',
          sourceTradition: '',
        })
        setShowDetails(false)
      })
    },
  })

  const reflection = watch('reflection')
  const passage = watch('passage')
  const sourceTitle = watch('sourceTitle')
  const sourceAuthor = watch('sourceAuthor')
  const mood = watch('mood')
  const tagsField = watch('tags')
  const deferredReflection = useDeferredValue(reflection)

  const parsedTags = splitTags(tagsField)
  const isUnauthorized = recentEntriesQuery.isError && isApiError(recentEntriesQuery.error) && recentEntriesQuery.error.status === 401

  const onSubmit = handleSubmit(async (values) => {
    setSavedEntry(null)
    await createMutation.mutateAsync(values)
  })

  return (
    <section className="page">
      <section className="hero-grid">
        <article className="hero-card hero-card-feature">
          <p className="eyebrow">Capture desk</p>
          <h2>Record the note before you optimize it.</h2>
          <p className="hero-copy">
            Quick mode handles source, passage, and reflection. Open the study details only when tags or mood actually sharpen the note.
          </p>
          <div className="hero-actions">
            <button
              type="button"
              className="button button-secondary"
              onClick={() => setShowDetails((current) => !current)}
            >
              {showDetails ? 'Hide study details' : 'Show study details'}
            </button>
          </div>
        </article>

        <article className="hero-card hero-card-note">
          <p className="eyebrow">Current posture</p>
          <p className="hero-copy">
            The right note is usually the one that preserves your own language, not the one that sounds finished.
          </p>
        </article>
      </section>

      <section className="composer-grid">
        <article className="panel panel-tall">
          <div className="panel-header">
            <div>
              <p className="eyebrow">New entry</p>
              <h3>Quick capture</h3>
            </div>
            <span className="meta-pill">Required first</span>
          </div>

          {isUnauthorized ? (
            <div className="inline-callout">
              <p className="error">You need an active session before you can save entries.</p>
              <Link className="text-link" to="/login">
                Sign in
              </Link>
            </div>
          ) : null}

          <form className="entry-form" onSubmit={onSubmit}>
            <label>
              Source title
              <input
                type="text"
                placeholder="The Cloud of Unknowing"
                {...register('sourceTitle', { required: 'Source title is required' })}
              />
            </label>

            <div className="field-row">
              <label>
                Author
                <input type="text" placeholder="Anonymous" {...register('sourceAuthor')} />
              </label>

              <label>
                Tradition
                <input type="text" placeholder="Christian" {...register('sourceTradition')} />
              </label>
            </div>

            <label>
              Passage
              <textarea rows={4} placeholder="A phrase, sentence, or short passage" {...register('passage')} />
            </label>

            <label>
              Reflection
              <textarea
                rows={10}
                placeholder="What did this reading disturb, reveal, or clarify?"
                {...register('reflection', {
                  required: 'Reflection is required',
                  maxLength: {
                    value: 10000,
                    message: 'Reflection must stay under 10000 characters',
                  },
                })}
              />
            </label>

            {errors.sourceTitle && <p className="error">{errors.sourceTitle.message}</p>}
            {errors.reflection && <p className="error">{errors.reflection.message}</p>}

            <button
              type="button"
              className="disclosure"
              onClick={() => setShowDetails((current) => !current)}
            >
              {showDetails ? 'Study details are open' : 'Add tags, mood, and energy'}
            </button>

            {showDetails ? (
              <section className="detail-card">
                <div className="field-row field-row-tight">
                  <label>
                    Mood
                    <input type="text" placeholder="focused" {...register('mood')} />
                  </label>

                  <label>
                    Energy
                    <select {...register('energy')}>
                      <option value="">Unspecified</option>
                      <option value="1">1</option>
                      <option value="2">2</option>
                      <option value="3">3</option>
                      <option value="4">4</option>
                      <option value="5">5</option>
                    </select>
                  </label>
                </div>

                <label>
                  Tags
                  <input type="text" placeholder="kenosis, prayer, surrender" {...register('tags')} />
                </label>

                {parsedTags.length ? (
                  <div className="tag-row">
                    {parsedTags.map((tag) => (
                      <span key={tag} className="tag-pill">
                        {tag}
                      </span>
                    ))}
                  </div>
                ) : (
                  <p className="muted-copy compact-copy">Separate tags with commas. Duplicates collapse automatically.</p>
                )}
              </section>
            ) : null}

            {createMutation.isError ? (
              <p className="error">{createMutation.error instanceof Error ? createMutation.error.message : 'Could not save entry'}</p>
            ) : null}
            {savedEntry ? (
              <p className="success">Entry #{savedEntry.id} saved. Capture another while the thread is still open.</p>
            ) : null}

            <div className="submit-row">
              <button type="submit" disabled={isSubmitting || createMutation.isPending || isUnauthorized}>
                {createMutation.isPending ? 'Saving entry...' : 'Save entry'}
              </button>
              <p className="muted-copy compact-copy">Source titles are created on submit if they do not exist yet.</p>
            </div>
          </form>
        </article>

        <aside className="composer-sidebar">
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Preview</p>
                <h3>Reading surface</h3>
              </div>
            </div>
            <div className="preview-meta">
              <span className="meta-pill">{sourceTitle.trim() || 'Untitled source'}</span>
              {sourceAuthor.trim() ? <span className="meta-pill">{sourceAuthor.trim()}</span> : null}
              {mood.trim() ? <span className="meta-pill">{mood.trim()}</span> : null}
            </div>
            {passage.trim() ? <blockquote className="passage-block">{passage.trim()}</blockquote> : null}
            <div className="markdown-preview">
              {deferredReflection.trim() ? (
                <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
                  {deferredReflection}
                </ReactMarkdown>
              ) : (
                <p className="muted-copy">Your reflection preview appears here as you write.</p>
              )}
            </div>
          </article>

          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Recent notes</p>
                <h3>Keep continuity</h3>
              </div>
            </div>
            {recentEntriesQuery.isPending ? <p className="muted-copy">Loading recent entries...</p> : null}
            {recentEntriesQuery.isError && !isUnauthorized ? <p className="error">Could not load recent entries.</p> : null}
            {isUnauthorized ? (
              <p className="muted-copy">Recent entries will appear after sign-in.</p>
            ) : (
              <div className="entry-list">
                {(recentEntriesQuery.data?.data ?? []).map((entry) => (
                  <article key={entry.id} className="entry-card entry-card-compact">
                    <div className="entry-card-header">
                      <span className="meta-pill">#{entry.id}</span>
                      <time>{formatShortDate(entry.created_at)}</time>
                    </div>
                    <p className="entry-excerpt">{excerpt(entry.reflection, 110)}</p>
                  </article>
                ))}
              </div>
            )}
          </article>
        </aside>
      </section>
    </section>
  )
}

function splitTags(value: string) {
  const seen = new Set<string>()
  const tags: string[] = []

  for (const part of value.split(',')) {
    const normalized = part.trim()
    const key = normalized.toLowerCase()
    if (!normalized || seen.has(key)) {
      continue
    }
    seen.add(key)
    tags.push(normalized)
  }

  return tags
}

function excerpt(value: string, limit: number) {
  if (value.length <= limit) {
    return value
  }
  return `${value.slice(0, limit).trimEnd()}...`
}

function formatShortDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
  })
}
