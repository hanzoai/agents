'use client'

import Image from 'next/image'
import { Anchor, Text, XStack } from '@hanzo/ui'
import type { ComponentProps } from 'react'

export function PoweredBy(props: ComponentProps<typeof XStack>) {
  return (
    <XStack {...props}>
      <Anchor
        href="https://github.com/hanzoai/agents"
        target="_blank"
        rel="noopener noreferrer"
        textDecorationLine="none"
      >
        <XStack alignItems="center" gap="$2" opacity={0.8} hoverStyle={{ opacity: 1 }}>
          <Text fontSize={12} color="$color11">
            Powered by
          </Text>
          <Image src="/hanzo-agents-logo-dark.svg" alt="Hanzo Agents" width={80} height={16} />
        </XStack>
      </Anchor>
    </XStack>
  )
}
