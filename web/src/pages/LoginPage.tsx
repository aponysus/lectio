import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { apiFetch } from '../api/client'

type LoginForm = {
  email: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>()
  const [message, setMessage] = useState<string>('')
  const [error, setError] = useState<string>('')

  const loginMutation = useMutation({
    mutationFn: (values: LoginForm) =>
      apiFetch<void>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(values),
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries()
      setMessage('Session established.')
      navigate('/entries/new')
    },
  })

  const onSubmit = handleSubmit(async (values) => {
    setMessage('')
    setError('')
    try {
      await loginMutation.mutateAsync(values)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      setError(errorMessage)
    }
  })

  return (
    <section className="page page-single">
      <div className="login-layout">
        <article className="hero-card hero-card-note">
          <p className="eyebrow">Private workspace</p>
          <h2>Enter the reading room.</h2>
          <p className="hero-copy">
            Lectio is designed as a single-user study notebook. Once signed in, entry capture and review stay in the same quiet space.
          </p>
          <ul className="feature-list">
            <li>Quick capture for source, passage, and reflection</li>
            <li>Markdown rendering for note review</li>
            <li>Tag-driven revisiting once your note history grows</li>
          </ul>
          <Link className="text-link" to="/entries/new">
            Go straight to the entry desk
          </Link>
        </article>

        <article className="panel login-panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Authentication</p>
              <h3>Sign in</h3>
            </div>
          </div>

          <form className="form" onSubmit={onSubmit}>
            <label>
              Email
              <input type="email" autoComplete="username" {...register('email')} />
            </label>

            <label>
              Password
              <input
                type="password"
                autoComplete="current-password"
                {...register('password', { required: 'Password is required' })}
              />
            </label>

            {errors.password && <p className="error">{errors.password.message}</p>}
            {error && <p className="error">{error}</p>}
            {message && <p className="success">{message}</p>}

            <button type="submit" disabled={isSubmitting || loginMutation.isPending}>
              {isSubmitting || loginMutation.isPending ? 'Signing in...' : 'Sign in'}
            </button>
          </form>
        </article>
      </div>
    </section>
  )
}
