'use client'

import Image from 'next/image'
import { Text, XStack, YStack } from '@hanzo/ui'
import { useEvaluation } from '@/hooks/useEvaluation'
import { EvaluationInput } from '@/lib/types'
import { DashboardShell, DashboardHeader, DashboardMain } from '@/components/eval/DashboardShell'
import { InputPanel } from '@/components/eval/InputPanel'
import { ResultsPanel } from '@/components/eval/ResultsPanel'
import { PoweredBy } from '@/components/PoweredBy'

export default function Home() {
  const { status, result, error, notes, currentStep, evaluate, reset } = useEvaluation()

  const handleSubmit = async (input: EvaluationInput) => {
    await evaluate(input)
  }

  return (
    <DashboardShell>
      <DashboardHeader>
        <XStack alignItems="center" gap="$3">
          <YStack
            height={32}
            width={32}
            borderRadius="$4"
            borderWidth={1}
            borderColor="$borderColor"
            backgroundColor="$color2"
            alignItems="center"
            justifyContent="center"
          >
            <Image src="/hanzo-agents-icon-dark.svg" alt="Hanzo Agents" width={16} height={16} />
          </YStack>
          <YStack>
            <Text fontSize={14} fontWeight="600">
              RAG Evaluation Studio
            </Text>
            <Text fontSize={10} fontFamily="$mono" color="$color11">
              v0.1.0-beta
            </Text>
          </YStack>
        </XStack>
        <XStack alignItems="center" gap="$6">
          <XStack borderLeftWidth={1} borderLeftColor="$borderColor" paddingLeft="$6" display="none" $md={{ display: 'flex' }}>
            <PoweredBy />
          </XStack>
        </XStack>
      </DashboardHeader>

      <DashboardMain flexDirection="column" $md={{ flexDirection: 'row' }}>
        <InputPanel
          onSubmit={handleSubmit}
          isLoading={status === 'evaluating'}
          width="100%"
          $md={{ width: 420 }}
          $lg={{ width: 480 }}
          flexShrink={0}
          height="100%"
          overflow="hidden"
        />
        <ResultsPanel
          status={status}
          result={result}
          error={error}
          notes={notes}
          currentStep={currentStep}
          onReset={reset}
          flex={1}
          minWidth={0}
          height="100%"
          overflow="hidden"
        />
      </DashboardMain>
    </DashboardShell>
  )
}
