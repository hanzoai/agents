import type { Metadata } from 'next'
import { Hanzo } from '@hanzo/ui'
import './globals.css'

export const metadata: Metadata = {
  title: 'RAG Evaluator',
  description: 'Multi-perspective evaluation for RAG-generated responses',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <Hanzo>{children}</Hanzo>
      </body>
    </html>
  )
}
