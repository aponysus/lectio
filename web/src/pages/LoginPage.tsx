import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { useSession } from '../hooks/useSession'
import { formFieldOnDarkClassName } from '../components/shared/formStyles'

const loginSchema = z.object({
  password: z.string().min(1, 'Password is required'),
})

type LoginForm = z.infer<typeof loginSchema>

export function LoginPage() {
  const navigate = useNavigate()
  const { login } = useSession()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (values: LoginForm) => {
    try {
      await login(values.password)
      navigate('/', { replace: true })
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Login failed'
      setError('password', { message })
    }
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_right,_rgba(39,71,59,0.22),_transparent_30%),linear-gradient(180deg,_#f4efe4_0%,_#e8dcc5_100%)] px-6 py-10 text-ink">
      <div className="mx-auto grid max-w-6xl gap-8 lg:grid-cols-[1.2fr_0.8fr]">
        <section className="rounded-[2.5rem] border border-black/5 bg-white/70 p-8 shadow-card backdrop-blur lg:p-12">
          <p className="text-xs uppercase tracking-[0.35em] text-accent/80">Lectio</p>
          <h1 className="mt-4 max-w-2xl font-display text-5xl leading-tight text-ink">A private workspace for serious source work.</h1>
          <p className="mt-6 max-w-xl text-base leading-7 text-ink/72">
            Log encounters, connect them to live inquiries, extract claims, revisit unresolved material, and compress it into synthesis.
          </p>
          <div className="mt-10 grid gap-4 sm:grid-cols-3">
            <div className="rounded-2xl border border-black/5 bg-canvas/70 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/80">Loop</p>
              <p className="mt-2 text-sm text-ink/80">Source to engagement to inquiry to claim to synthesis.</p>
            </div>
            <div className="rounded-2xl border border-black/5 bg-canvas/70 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/80">Stack</p>
              <p className="mt-2 text-sm text-ink/80">Go, chi, SQLite, React, Vite, Docker, Terraform.</p>
            </div>
            <div className="rounded-2xl border border-black/5 bg-canvas/70 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/80">Now</p>
              <p className="mt-2 text-sm text-ink/80">Resume current work, capture the next encounter, or reopen what still needs thought.</p>
            </div>
          </div>
        </section>

        <section className="rounded-[2rem] border border-black/5 bg-stone-950 px-6 py-8 text-stone-100 shadow-card lg:px-8">
          <p className="text-xs uppercase tracking-[0.25em] text-stone-400">Sign in</p>
          <h2 className="mt-3 font-display text-3xl">Enter the scaffold</h2>
          <form className="mt-8 space-y-4" onSubmit={handleSubmit(onSubmit)}>
            <label className="block">
              <span className="mb-2 block text-sm text-stone-300">Bootstrap password</span>
              <input
                {...register('password')}
                type="password"
                autoComplete="current-password"
                className={formFieldOnDarkClassName}
              />
              {errors.password ? <span className="mt-2 block text-sm text-amber-300">{errors.password.message}</span> : null}
            </label>
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90 disabled:cursor-wait disabled:opacity-70"
            >
              {isSubmitting ? 'Signing in...' : 'Sign in'}
            </button>
          </form>
          <p className="mt-6 text-sm leading-6 text-stone-400">
            Local development defaults are provided in <code>.env.example</code>. Change them before deploying anywhere
            real.
          </p>
        </section>
      </div>
    </div>
  )
}
