import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeSanitize from 'rehype-sanitize'

const sampleMarkdown = `# Welcome to Lectio

This scaffold includes:

- React Router for SPA routing
- TanStack Query for server-state
- React Hook Form for form handling
- react-markdown + GFM + sanitize for reflection rendering
`

export function DashboardPage() {
  return (
    <section className="panel">
      <h2>Dashboard</h2>
      <p>Daily resonance and timeline summaries will appear here.</p>
      <article className="markdown-preview">
        <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
          {sampleMarkdown}
        </ReactMarkdown>
      </article>
    </section>
  )
}
