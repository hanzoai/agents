'use client'

import { Loader2, CheckCircle, Circle, ArrowRight } from 'lucide-react'
import {
  Badge,
  Card,
  CardContent,
  ScrollView,
  Text,
  XStack,
  YStack,
} from '@hanzo/ui'
import type { BadgeVariant } from '@hanzo/ui'
import { WorkflowNote } from '@/lib/types'

interface LoadingStateProps {
  notes: WorkflowNote[]
  currentStep: string | null
}

function getTagVariant(tag: string): BadgeVariant {
  const variants: Record<string, BadgeVariant> = {
    orchestration: 'default',
    faithfulness: 'secondary',
    relevance: 'secondary',
    hallucination: 'secondary',
    constitutional: 'secondary',
    complete: 'default',
  }
  return variants[tag] || 'outline'
}

function formatTime(timestamp: string): string {
  const diffSec = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000)
  if (diffSec < 1) return 'now'
  if (diffSec < 60) return `${diffSec}s ago`
  return `${Math.floor(diffSec / 60)}m ago`
}

export function LoadingState({ notes, currentStep }: LoadingStateProps) {
  return (
    <YStack maxWidth={512} width="100%">
      <Card>
        <CardContent>
          <YStack paddingTop="$8" paddingBottom="$6" gap="$6">
            <YStack alignItems="center" gap="$2">
              <Loader2 size={40} />
              <Text fontSize={18} fontWeight="500">
                Evaluating your response...
              </Text>
              <Text fontSize={14} color="$color11">
                Multi-perspective analysis in progress
              </Text>
            </YStack>

            {currentStep && (
              <XStack alignItems="center" justifyContent="center" gap="$2" paddingVertical="$2" paddingHorizontal="$4" borderRadius="$4" backgroundColor="$color2">
                <ArrowRight size={16} />
                <Text fontSize={14} fontWeight="500" numberOfLines={1}>
                  {currentStep}
                </Text>
              </XStack>
            )}

            <YStack borderWidth={1} borderColor="$borderColor" borderRadius="$4">
              <XStack
                paddingHorizontal="$3"
                paddingVertical="$2"
                borderBottomWidth={1}
                borderBottomColor="$borderColor"
                alignItems="center"
                justifyContent="space-between"
                backgroundColor="$color2"
              >
                <Text fontSize={12} fontWeight="500" color="$color11" textTransform="uppercase">
                  Workflow Log
                </Text>
                <Text fontSize={12} color="$color11">
                  {notes.length} events
                </Text>
              </XStack>
              <ScrollView height={224}>
                <YStack padding="$2" gap="$1">
                  {notes.length === 0 ? (
                    <YStack alignItems="center" paddingVertical="$8" gap="$2">
                      <Loader2 size={20} opacity={0.5} />
                      <Text fontSize={14} color="$color11">
                        Waiting for workflow events...
                      </Text>
                    </YStack>
                  ) : (
                    notes.map((note, index) => (
                      <XStack
                        key={`${note.timestamp}-${index}`}
                        alignItems="flex-start"
                        gap="$2"
                        paddingVertical={6}
                        paddingHorizontal="$2"
                        borderRadius="$2"
                      >
                        <YStack marginTop={2} flexShrink={0}>
                          {index === notes.length - 1 ? (
                            <Circle size={14} />
                          ) : (
                            <CheckCircle size={14} />
                          )}
                        </YStack>
                        <YStack flex={1} minWidth={0} gap="$1">
                          <Text
                            fontSize={14}
                            color={index === notes.length - 1 ? '$color' : '$color11'}
                          >
                            {note.message}
                          </Text>
                          <XStack alignItems="center" gap={6} flexWrap="wrap">
                            {note.tags.slice(0, 3).map((tag) => (
                              <Badge key={tag} variant={getTagVariant(tag)}>
                                {tag}
                              </Badge>
                            ))}
                            <Text fontSize={12} color="$color11">
                              {formatTime(note.timestamp)}
                            </Text>
                          </XStack>
                        </YStack>
                      </XStack>
                    ))
                  )}
                </YStack>
              </ScrollView>
            </YStack>

            <Text textAlign="center" fontSize={12} color="$color11">
              Typically takes 5-15 seconds
            </Text>
          </YStack>
        </CardContent>
      </Card>
    </YStack>
  )
}
