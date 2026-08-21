'use client'

import { AlertTriangle } from 'lucide-react'
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Progress,
  Text,
  XStack,
  YStack,
} from '@hanzo/ui'
import type { BadgeVariant } from '@hanzo/ui'
import { EvaluationResult } from '@/lib/types'

interface OverallScoreProps {
  result: EvaluationResult
}

function formatScore(score: number): string {
  return `${Math.round(score * 100)}%`
}

function getQualityVariant(tier: string): BadgeVariant {
  switch (tier) {
    case 'excellent':
    case 'good':
      return 'default'
    case 'acceptable':
      return 'secondary'
    case 'poor':
    case 'critical':
      return 'destructive'
    default:
      return 'outline'
  }
}

export function OverallScore({ result }: OverallScoreProps) {
  const { overallScore, qualityTier, faithfulness, relevance, hallucination, constitutional } =
    result

  const metrics = [
    { name: 'Faithfulness', score: faithfulness.score },
    { name: 'Relevance', score: relevance.score },
    { name: 'Hallucination', score: hallucination.score },
    { name: 'Constitutional', score: constitutional.score },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Overall Score</CardTitle>
      </CardHeader>
      <CardContent>
        <YStack gap="$6">
          <YStack gap="$1">
            <XStack alignItems="baseline" gap="$3">
              <Text fontSize={36} fontWeight="700">
                {formatScore(overallScore)}
              </Text>
              <Badge variant={getQualityVariant(qualityTier)}>{qualityTier}</Badge>
            </XStack>
            <Text fontSize={14} color="$color11">
              Quality assessment of the RAG response
            </Text>
          </YStack>

          <YStack gap="$3">
            {metrics.map((metric) => (
              <YStack key={metric.name} gap={6}>
                <XStack alignItems="center" justifyContent="space-between">
                  <Text fontSize={14} color="$color11">
                    {metric.name}
                  </Text>
                  <Text fontSize={14} fontFamily="$mono">
                    {formatScore(metric.score)}
                  </Text>
                </XStack>
                <Progress value={metric.score * 100} height={8} />
              </YStack>
            ))}
          </YStack>

          {result.recommendations && result.recommendations.length > 0 && (
            <YStack borderRadius="$4" borderWidth={1} borderColor="$borderColor" padding="$4" gap="$2">
              <XStack alignItems="center" gap="$2">
                <AlertTriangle size={16} />
                <Text fontSize={14} fontWeight="500">
                  Recommendations
                </Text>
              </XStack>
              <YStack gap="$1">
                {result.recommendations.slice(0, 3).map((rec, i) => (
                  <Text key={i} fontSize={14} color="$color11">
                    {'• '}
                    {rec}
                  </Text>
                ))}
              </YStack>
            </YStack>
          )}
        </YStack>
      </CardContent>
    </Card>
  )
}
