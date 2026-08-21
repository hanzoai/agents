'use client'

import type { ComponentProps } from 'react'
import { RotateCcw, AlertCircle, FileJson, BarChart2, ListChecks, Activity } from 'lucide-react'
import {
  Button,
  ScrollView,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Text,
  XStack,
  YStack,
} from '@hanzo/ui'
import { EvaluationResult, WorkflowNote } from '@/lib/types'
import { LoadingState } from '@/components/LoadingState'
import { MetricCard } from '@/components/MetricCard'
import { OverallScore } from '@/components/OverallScore'
import { ClaimBreakdown } from '@/components/ClaimBreakdown'

type ResultsPanelProps = ComponentProps<typeof YStack> & {
  status: 'idle' | 'evaluating' | 'success' | 'error'
  result: EvaluationResult | null
  error: string | null
  notes: WorkflowNote[]
  currentStep: string | null
  onReset: () => void
}

const TAB_TRIGGERS = [
  { value: 'metrics', label: 'Metrics Overview', Icon: BarChart2 },
  { value: 'claims', label: 'Claim Analysis', Icon: ListChecks },
  { value: 'json', label: 'Raw Output', Icon: FileJson },
] as const

export function ResultsPanel({
  status,
  result,
  error,
  notes,
  currentStep,
  onReset,
  ...frame
}: ResultsPanelProps) {
  if (status === 'idle') {
    return (
      <YStack alignItems="center" justifyContent="center" padding="$8" backgroundColor="$background" {...frame}>
        <YStack
          width={80}
          height={80}
          borderRadius={9999}
          backgroundColor="$color3"
          alignItems="center"
          justifyContent="center"
          marginBottom="$6"
          borderWidth={1}
          borderColor="$borderColor"
        >
          <Activity size={40} opacity={0.4} />
        </YStack>
        <Text fontSize={20} fontWeight="500">
          Ready to Evaluate
        </Text>
        <Text maxWidth={384} marginTop="$3" fontSize={14} color="$color11" textAlign="center">
          Configure your evaluation parameters and inputs on the left, then click &ldquo;Run
          Evaluation&rdquo;.
        </Text>
      </YStack>
    )
  }

  if (status === 'evaluating') {
    return (
      <YStack padding="$8" alignItems="center" justifyContent="center" backgroundColor="$background" {...frame}>
        <LoadingState notes={notes} currentStep={currentStep} />
      </YStack>
    )
  }

  if (status === 'error') {
    return (
      <YStack padding="$8" alignItems="center" justifyContent="center" backgroundColor="$background" {...frame}>
        <YStack
          maxWidth={448}
          width="100%"
          borderWidth={1}
          borderColor="$destructive"
          borderRadius="$6"
          padding="$8"
          alignItems="center"
          gap="$4"
        >
          <YStack width={48} height={48} borderRadius={9999} backgroundColor="$color3" alignItems="center" justifyContent="center">
            <AlertCircle size={24} color="var(--destructive)" />
          </YStack>
          <Text fontSize={18} fontWeight="600" color="$destructive">
            Evaluation Failed
          </Text>
          <Text fontSize={14} color="$color11" textAlign="center">
            {error || 'An unknown error occurred'}
          </Text>
          <Button variant="outline" onPress={onReset} minWidth={120}>
            Try Again
          </Button>
        </YStack>
      </YStack>
    )
  }

  return (
    <YStack backgroundColor="$background" minWidth={0} {...frame}>
      <XStack
        height={64}
        alignItems="center"
        justifyContent="space-between"
        paddingHorizontal="$6"
        borderBottomWidth={1}
        borderBottomColor="$borderColor"
        flexShrink={0}
        zIndex={10}
      >
        <Text fontSize={18} fontWeight="600">
          Evaluation Report
        </Text>
        <Button variant="ghost" size="sm" onPress={onReset} gap="$2">
          <RotateCcw size={16} />
          <Text fontSize={13}>New Evaluation</Text>
        </Button>
      </XStack>

      <Tabs defaultValue="metrics" flex={1} flexDirection="column" minWidth={0}>
        <XStack paddingHorizontal="$6" borderBottomWidth={1} borderBottomColor="$borderColor" flexShrink={0}>
          <TabsList backgroundColor="transparent" gap="$6" height={48}>
            {TAB_TRIGGERS.map(({ value, label, Icon }) => (
              <TabsTrigger key={value} value={value} backgroundColor="transparent" paddingHorizontal={0} height={48} gap="$2">
                <Icon size={16} />
                <Text fontSize={14}>{label}</Text>
              </TabsTrigger>
            ))}
          </TabsList>
        </XStack>

        <ScrollView flex={1} minWidth={0}>
          <YStack maxWidth={1152} width="100%" marginHorizontal="auto" padding="$8" minWidth={0}>
            <TabsContent value="metrics">
              {result && (
                <YStack gap="$8">
                  <OverallScore result={result} />
                  <YStack gap="$4">
                    <Text fontSize={18} fontWeight="600">
                      Detailed Metrics
                    </Text>
                    <XStack gap="$6" flexWrap="wrap">
                      <MetricCard metric="faithfulness" result={result} />
                      <MetricCard metric="relevance" result={result} />
                      <MetricCard metric="hallucination" result={result} />
                      <MetricCard metric="constitutional" result={result} />
                    </XStack>
                  </YStack>
                </YStack>
              )}
            </TabsContent>

            <TabsContent value="claims">{result && <ClaimBreakdown result={result} />}</TabsContent>

            <TabsContent value="json">
              {result && (
                <YStack borderRadius="$6" borderWidth={1} borderColor="$borderColor" backgroundColor="$card" overflow="hidden">
                  <ScrollView maxHeight={520} width="100%">
                    <YStack padding="$4">
                      <Text fontFamily="$mono" fontSize={12} color="$color11">
                        {JSON.stringify(result, null, 2)}
                      </Text>
                    </YStack>
                  </ScrollView>
                </YStack>
              )}
            </TabsContent>
          </YStack>
        </ScrollView>
      </Tabs>
    </YStack>
  )
}
