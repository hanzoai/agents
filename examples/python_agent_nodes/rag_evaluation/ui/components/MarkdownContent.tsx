'use client'

import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Anchor, Text, YStack } from '@hanzo/ui'
import type { ComponentProps, ReactNode } from 'react'

type MarkdownContentProps = ComponentProps<typeof YStack> & {
  content: string
}

const heading = (fontSize: number, fontWeight: '500' | '600', marginTop: string) =>
  function Heading({ children }: { children?: ReactNode }) {
    return (
      <Text fontSize={fontSize} fontWeight={fontWeight} marginTop={marginTop} marginBottom="$2">
        {children}
      </Text>
    )
  }

const body = ({ children }: { children?: ReactNode }) => (
  <Text fontSize={14} color="$color11" marginBottom="$2">
    {children}
  </Text>
)

const listItem = ({ children }: { children?: ReactNode }) => (
  <Text fontSize={14} color="$color11">
    {'• '}
    {children}
  </Text>
)

export function MarkdownContent({ content, ...frame }: MarkdownContentProps) {
  if (!content) return null

  return (
    <YStack {...frame}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: heading(18, '600', '$4'),
          h2: heading(16, '600', '$4'),
          h3: heading(14, '600', '$3'),
          h4: heading(14, '500', '$2'),
          p: body,
          ul: ({ children }) => (
            <YStack paddingLeft="$4" gap="$1" marginBottom="$2">
              {children}
            </YStack>
          ),
          ol: ({ children }) => (
            <YStack paddingLeft="$4" gap="$1" marginBottom="$2">
              {children}
            </YStack>
          ),
          li: listItem,
          strong: ({ children }) => <Text fontWeight="600">{children}</Text>,
          em: ({ children }) => (
            <Text fontStyle="italic" color="$color11">
              {children}
            </Text>
          ),
          code: ({ children }) => (
            <Text backgroundColor="$color3" paddingHorizontal={6} paddingVertical={2} borderRadius="$2" fontSize={12} fontFamily="$mono">
              {children}
            </Text>
          ),
          pre: ({ children }) => (
            <YStack backgroundColor="$color3" padding="$3" borderRadius="$4" marginBottom="$2">
              <Text fontSize={12} fontFamily="$mono">
                {children}
              </Text>
            </YStack>
          ),
          blockquote: ({ children }) => (
            <YStack borderLeftWidth={2} borderLeftColor="$primary" paddingLeft="$3" marginBottom="$2">
              <Text fontStyle="italic" color="$color11">
                {children}
              </Text>
            </YStack>
          ),
          hr: () => <YStack borderTopWidth={1} borderTopColor="$borderColor" marginVertical="$4" />,
          a: ({ href, children }) => (
            <Anchor href={href} target="_blank" rel="noopener noreferrer" color="$primary">
              {children}
            </Anchor>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </YStack>
  )
}
