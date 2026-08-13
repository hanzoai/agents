import { XStack, YStack } from '@hanzo/ui'
import type { ComponentProps } from 'react'

export function DashboardShell(props: ComponentProps<typeof YStack>) {
  return <YStack flex={1} height="100vh" overflow="hidden" backgroundColor="$background" {...props} />
}

export function DashboardHeader(props: ComponentProps<typeof XStack>) {
  return (
    <XStack
      height={56}
      alignItems="center"
      justifyContent="space-between"
      paddingHorizontal="$6"
      backgroundColor="$background"
      borderBottomWidth={1}
      borderBottomColor="$borderColor"
      flexShrink={0}
      zIndex={10}
      {...props}
    />
  )
}

export function DashboardMain(props: ComponentProps<typeof XStack>) {
  return <XStack flex={1} overflow="hidden" {...props} />
}
