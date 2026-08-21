'use client'

import { Scale, Target, Search, ScrollText, Check, AlertTriangle, X } from 'lucide-react'
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  Progress,
  ScrollView,
  Separator,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Text,
  XStack,
  YStack,
} from '@hanzo/ui'
import { EvaluationResult } from '@/lib/types'
import { MarkdownContent } from './MarkdownContent'

interface MetricCardProps {
  metric: 'faithfulness' | 'relevance' | 'hallucination' | 'constitutional'
  result: EvaluationResult
}

const metricConfig = {
  faithfulness: { icon: Scale, title: 'Faithfulness', subtitle: 'Adversarial Debate Pattern' },
  relevance: { icon: Target, title: 'Relevance', subtitle: 'Multi-Jury Consensus' },
  hallucination: { icon: Search, title: 'Hallucination', subtitle: 'Hybrid ML + LLM Verification' },
  constitutional: {
    icon: ScrollText,
    title: 'Constitutional',
    subtitle: 'Principles-Based Compliance',
  },
}

function formatScore(score: number): string {
  return `${Math.round(score * 100)}%`
}

function Stat({ value, label, tone }: { value: number; label: string; tone?: 'bad' | 'good' }) {
  return (
    <YStack flex={1} gap="$1" alignItems="center">
      <Text
        fontSize={24}
        fontWeight="700"
        color={tone === 'bad' ? '$destructive' : tone === 'good' ? '$primary' : '$color'}
      >
        {value}
      </Text>
      <Text fontSize={12} color="$color11">
        {label}
      </Text>
    </YStack>
  )
}

function Row({ label, value, tone }: { label: string; value: number; tone?: 'bad' | 'good' }) {
  return (
    <XStack alignItems="center" justifyContent="space-between">
      <Text fontSize={14} color="$color11">
        {label}
      </Text>
      <Text
        fontSize={14}
        fontFamily="$mono"
        color={tone === 'bad' ? '$destructive' : tone === 'good' ? '$primary' : '$color'}
      >
        {value}
      </Text>
    </XStack>
  )
}

export function MetricCard({ metric, result }: MetricCardProps) {
  const config = metricConfig[metric]
  const Icon = config.icon
  const data = result[metric]

  return (
    <Card flexGrow={1} flexBasis="45%" minWidth={320}>
      <CardHeader>
        <XStack alignItems="flex-start" justifyContent="space-between">
          <XStack alignItems="center" gap="$3">
            <YStack padding="$2" borderRadius="$4" backgroundColor="$color3">
              <Icon size={20} />
            </YStack>
            <YStack>
              <Text fontWeight="500">{config.title}</Text>
              <Text fontSize={12} color="$color11">
                {config.subtitle}
              </Text>
            </YStack>
          </XStack>
          <Text fontSize={24} fontWeight="700">
            {formatScore(data.score)}
          </Text>
        </XStack>
      </CardHeader>

      <Separator />

      <CardContent>
        <Tabs defaultValue="summary" flexDirection="column" paddingTop="$4">
          <TabsList width="100%" marginBottom="$4">
            <TabsTrigger value="summary" flex={1}>
              <Text fontSize={13}>Summary</Text>
            </TabsTrigger>
            <TabsTrigger value="details" flex={1}>
              <Text fontSize={13}>Details</Text>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="summary" flexBasis="auto">
            {metric === 'faithfulness' && (
              <XStack gap="$3">
                <Stat value={result.faithfulness.prosecutorIssues} label="Issues Found" />
                <Stat value={result.faithfulness.defenderUpheld} label="Claims Upheld" />
                <Stat
                  value={result.faithfulness.unfaithfulClaims}
                  label="Unfaithful"
                  tone={result.faithfulness.unfaithfulClaims > 0 ? 'bad' : 'good'}
                />
              </XStack>
            )}

            {metric === 'relevance' && (
              <YStack gap="$3">
                {(['literal', 'intent', 'scope'] as const).map((type) => {
                  const score = result.relevance[
                    `${type}Score` as keyof typeof result.relevance
                  ] as number
                  return (
                    <YStack key={type} gap={6}>
                      <XStack alignItems="center" justifyContent="space-between">
                        <Text fontSize={14} color="$color11" textTransform="capitalize">
                          {type}
                        </Text>
                        <Text fontSize={14} fontFamily="$mono">
                          {formatScore(score)}
                        </Text>
                      </XStack>
                      <Progress value={score * 100} height={8} />
                    </YStack>
                  )
                })}
              </YStack>
            )}

            {metric === 'hallucination' && (
              <YStack gap="$3">
                <XStack gap="$3">
                  <YStack flex={1} padding="$3" borderRadius="$4" backgroundColor="$color2">
                    <Stat value={result.hallucination.statementsAnalyzed} label="Statements" />
                  </YStack>
                  <YStack flex={1} padding="$3" borderRadius="$4" backgroundColor="$color2">
                    <Stat value={result.hallucination.mlVerified} label="ML Verified" />
                  </YStack>
                </XStack>
                <Row label="LLM Escalated" value={result.hallucination.llmEscalated} />
                <Row
                  label="Fabrications Found"
                  value={result.hallucination.fabricationsFound}
                  tone={result.hallucination.fabricationsFound > 0 ? 'bad' : 'good'}
                />
              </YStack>
            )}

            {metric === 'constitutional' && (
              <YStack gap="$2">
                {result.constitutional.principles.slice(0, 5).map((principle) => (
                  <XStack key={principle.id} alignItems="center" justifyContent="space-between" paddingVertical="$1">
                    <Text fontSize={14} color="$color11">
                      {principle.name.replace(/_/g, ' ')}
                    </Text>
                    <Badge
                      variant={
                        principle.passed
                          ? 'default'
                          : principle.score >= 0.5
                            ? 'secondary'
                            : 'destructive'
                      }
                    >
                      {principle.passed ? (
                        <Check size={12} />
                      ) : principle.score >= 0.5 ? (
                        <AlertTriangle size={12} />
                      ) : (
                        <X size={12} />
                      )}
                      {principle.passed ? 'passed' : principle.score >= 0.5 ? 'issue' : 'failed'}
                    </Badge>
                  </XStack>
                ))}
              </YStack>
            )}
          </TabsContent>

          <TabsContent value="details" flexBasis="auto">
            <ScrollView maxHeight={192}>
              {metric === 'faithfulness' && (
                <YStack>
                  <Text fontSize={14} fontWeight="500" marginBottom="$2">
                    Debate Summary
                  </Text>
                  {result.faithfulness.debateSummary ? (
                    <MarkdownContent content={result.faithfulness.debateSummary} />
                  ) : (
                    <Text fontSize={14} color="$color11">
                      No debate summary available
                    </Text>
                  )}
                </YStack>
              )}

              {metric === 'relevance' && (
                <YStack>
                  <Text fontSize={14} fontWeight="500" marginBottom="$2">
                    Jury Verdict
                  </Text>
                  {result.relevance.verdict ? (
                    <MarkdownContent content={result.relevance.verdict} />
                  ) : (
                    <Text fontSize={14} color="$color11">
                      No verdict summary available
                    </Text>
                  )}
                  {result.relevance.disagreementLevel > 0.3 && (
                    <YStack marginTop="$3" padding="$2" borderRadius="$2" borderWidth={1} borderColor="$borderColor">
                      <Text fontSize={12}>
                        Note: Jury had {formatScore(result.relevance.disagreementLevel)} disagreement
                      </Text>
                    </YStack>
                  )}
                </YStack>
              )}

              {metric === 'hallucination' && (
                <YStack>
                  <Text fontSize={14} fontWeight="500" marginBottom="$2">
                    Verification Process
                  </Text>
                  <Text fontSize={14} color="$color11">
                    The hybrid verification process first used ML models (embeddings + NLI) to
                    verify {result.hallucination.mlVerified} statements.
                    {result.hallucination.llmEscalated > 0
                      ? ` ${result.hallucination.llmEscalated} uncertain cases were escalated to LLM for deeper analysis.`
                      : ''}
                  </Text>
                  {result.hallucination.fabricationsFound > 0 && (
                    <YStack marginTop="$3" padding="$2" borderRadius="$2" borderWidth={1} borderColor="$destructive">
                      <Text fontSize={12} color="$destructive">
                        {result.hallucination.fabricationsFound} fabrication(s) detected
                      </Text>
                    </YStack>
                  )}
                </YStack>
              )}

              {metric === 'constitutional' && (
                <YStack>
                  {result.constitutional.criticalViolations &&
                  result.constitutional.criticalViolations.length > 0 ? (
                    <YStack gap="$2">
                      <Text fontSize={14} fontWeight="500" color="$destructive">
                        Critical Violations
                      </Text>
                      {result.constitutional.criticalViolations.map((v, i) => (
                        <XStack key={i} alignItems="flex-start" gap="$2">
                          <X size={16} />
                          <Text fontSize={14} color="$color11" flex={1}>
                            {typeof v === 'string' ? v : JSON.stringify(v)}
                          </Text>
                        </XStack>
                      ))}
                    </YStack>
                  ) : (
                    <YStack gap="$2" alignItems="flex-start">
                      <Text fontSize={14} fontWeight="500">
                        Compliance Status
                      </Text>
                      <Badge
                        variant={
                          result.constitutional.complianceStatus === 'compliant'
                            ? 'default'
                            : result.constitutional.complianceStatus === 'minor_issues'
                              ? 'secondary'
                              : 'destructive'
                        }
                      >
                        {result.constitutional.complianceStatus.replace(/_/g, ' ')}
                      </Badge>
                      <Text fontSize={14} color="$color11">
                        All constitutional principles have been evaluated.{' '}
                        {result.constitutional.principles.filter((p) => p.passed).length} of{' '}
                        {result.constitutional.principles.length} principles passed.
                      </Text>
                    </YStack>
                  )}
                </YStack>
              )}
            </ScrollView>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
