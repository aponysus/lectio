import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { apiFetch } from '../api/client'

type LoginForm = {
  email: string
  password: string
}

export function LoginPage() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>()
  const [message, setMessage] = useState<string>('')
  const [error, setError] = useState<string>('')

  const onSubmit = handleSubmit(async (values) => {
    setMessage('')
    setError('')
    try {
      await apiFetch<void>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(values),
      })
      setMessage('Login succeeded. Session and CSRF cookies set.')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      setError(errorMessage)
    }
  })

  return (
    <section className="panel">
      <h2>Login</h2>
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

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
    </section>
  )
}
