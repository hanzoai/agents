'use client'

import { useState } from 'react'
import { ChevronDown, Gavel } from 'lucide-react'
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
  Separator,
  Text,
  XStack,
  YStack,
} from '@hanzo/ui'
import type { BadgeVariant } from '@hanzo/ui'
import { Claim, EvaluationResult } from '@/lib/types'

interface ClaimBreakdownProps {
  result: EvaluationResult
}

type FilterType = 'all' | 'grounded' | 'uncertain' | 'fabricated'

const FILTERS: { key: FilterType; label: string; on: BadgeVariant }[] = [
  { key: 'all', label: 'All', on: 'default' },
  { key: 'grounded', label: 'Grounded', on: 'default' },
  { key: 'uncertain', label: 'Uncertain', on: 'secondary' },
  { key: 'fabricated', label: 'Fabricated', on: 'destructive' },
]

const STATUS_VARIANT: Record<string, BadgeVariant> = {
  grounded: 'default',
  uncertain: 'secondary',
  fabricated: 'destructive',
}

export function ClaimBreakdown({ result }: ClaimBreakdownProps) {
  const [filter, setFilter] = useState<FilterType>('all')

  const claims = result.faithfulness.claims || []
  const filteredClaims = claims.filter((claim) => filter === 'all' || claim.status === filter)
  const getCount = (status: string) => claims.filter((c) => c.status === status).length

  return (
    <YStack gap="$6">
      <XStack alignItems="center" justifyContent="space-between" gap="$4" flexWrap="wrap">
        <YStack>
          <Text fontSize={18} fontWeight="500">
            Claims Analysis
          </Text>
          <Text fontSize={14} color="$color11" marginTop={6}>
            {claims.length} claims were identified and verified against the context.
          </Text>
        </YStack>

        <XStack alignItems="center" gap="$2">
          {FILTERS.map(({ key, label, on }) => (
            <Badge
              key={key}
              variant={filter === key ? on : 'outline'}
              onClick={() => setFilter(key)}
              style={{ cursor: 'pointer' }}
            >
              {label} {key === 'all' ? claims.length : getCount(key)}
            </Badge>
          ))}
        </XStack>
      </XStack>

      <YStack gap="$4">
        {filteredClaims.length === 0 ? (
          <YStack alignItems="center" paddingVertical="$8" borderWidth={1} borderColor="$borderColor" borderRadius="$4" borderStyle="dashed">
            <Text color="$color11">No claims found matching this filter.</Text>
          </YStack>
        ) : (
          filteredClaims.map((claim, index) => <ClaimItem key={claim.id || index} claim={claim} />)
        )}
      </YStack>
    </YStack>
  )
}

function Argument({
  title,
  badge,
  argument,
  kindLabel,
  kind,
}: {
  title: string
  badge?: React.ReactNode
  argument?: string
  kindLabel?: string
  kind?: string
}) {
  return (
    <Card flex={1} minWidth={280}>
      <CardHeader>
        <XStack alignItems="center" justifyContent="space-between">
          <CardTitle>{title}</CardTitle>
          {badge}
        </XStack>
      </CardHeader>
      <CardContent>
        {argument ? (
          <YStack gap="$4">
            <Text fontSize={14}>{argument}</Text>
            {kind && (
              <XStack alignItems="center" gap="$2">
                <Text fontSize={12} fontWeight="500" color="$color11">
                  {kindLabel}:
                </Text>
                <Text fontSize={12} backgroundColor="$color3" paddingHorizontal="$2" paddingVertical={2} borderRadius="$2">
                  {kind.replace(/_/g, ' ')}
                </Text>
              </XStack>
            )}
          </YStack>
        ) : (
          <Text fontSize={14} color="$color11" fontStyle="italic">
            No {title.toLowerCase()} arguments.
          </Text>
        )}
      </CardContent>
    </Card>
  )
}

function ClaimItem({ claim }: { claim: Claim }) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <Card overflow="hidden">
        <CollapsibleTrigger width="100%" alignItems="stretch">
          <YStack padding="$6" gap="$3" width="100%" hoverStyle={{ backgroundColor: '$color2' }}>
            <XStack alignItems="flex-start" justifyContent="space-between" gap="$4">
              <Text fontSize={14} fontWeight="500" flex={1} textAlign="left">
                &ldquo;{claim.text}&rdquo;
              </Text>
              <YStack rotate={isOpen ? '180deg' : '0deg'} marginTop={2}>
                <ChevronDown size={16} />
              </YStack>
            </XStack>

            <XStack alignItems="center" gap="$3" flexWrap="wrap">
              <Badge variant={STATUS_VARIANT[claim.status] || 'outline'}>{claim.status}</Badge>

              {claim.type && <Badge variant="outline">{claim.type.replace(/_/g, ' ')}</Badge>}

              {claim.evidence && (
                <Text fontSize={12} color="$color11" numberOfLines={1} maxWidth={300}>
                  Evidence: {claim.evidence}
                </Text>
              )}
            </XStack>
          </YStack>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <Separator />
          <YStack padding="$6" gap="$6">
            <XStack gap="$6" flexWrap="wrap">
              <Argument
                title="Prosecution"
                argument={claim.prosecution?.argument}
                kindLabel="Type"
                kind={claim.prosecution?.type}
                badge={
                  claim.prosecution ? (
                    <Badge variant="destructive">{claim.prosecution.severity}</Badge>
                  ) : undefined
                }
              />
              <Argument
                title="Defense"
                argument={claim.defense?.argument}
                kindLabel="Support"
                kind={claim.defense?.supportType}
                badge={
                  claim.defense ? (
                    <Badge variant="secondary">
                      Strength: {Math.round(claim.defense.strength * 100)}%
                    </Badge>
                  ) : undefined
                }
              />
            </XStack>

            <Card>
              <CardHeader>
                <XStack alignItems="center" justifyContent="space-between">
                  <XStack alignItems="center" gap="$2">
                    <Gavel size={16} />
                    <CardTitle>Judge Ruling</CardTitle>
                  </XStack>
                  {claim.judgeRuling && (
                    <Badge variant="outline">{claim.judgeRuling.verdict}</Badge>
                  )}
                </XStack>
              </CardHeader>
              <CardContent>
                <Text fontSize={14} color={claim.judgeRuling ? '$color' : '$color11'}>
                  {claim.judgeRuling ? claim.judgeRuling.reasoning : 'Pending ruling...'}
                </Text>
              </CardContent>
            </Card>
          </YStack>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}
