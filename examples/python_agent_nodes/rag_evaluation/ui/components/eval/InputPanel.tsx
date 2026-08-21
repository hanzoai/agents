'use client'

import { useState, useRef, useEffect } from 'react'
import type { ComponentProps } from 'react'
import { Play, ChevronDown, Check, Sparkles } from 'lucide-react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Input,
  Label,
  Popover,
  PopoverContent,
  PopoverTrigger,
  ScrollView,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Text,
  Textarea,
  XStack,
  YStack,
} from '@hanzo/ui'
import { EvaluationInput, AVAILABLE_MODELS } from '@/lib/types'
import { PRESETS, Preset } from '@/lib/presets'

type InputPanelProps = ComponentProps<typeof YStack> & {
  onSubmit: (input: EvaluationInput) => void
  isLoading: boolean
}

/** The catalogue is a literal tuple; a free-typed search box is compared against it as text. */
const knownModels: readonly string[] = AVAILABLE_MODELS

function RequiredTag() {
  return (
    <Text fontSize={10} color="$color11" backgroundColor="$color3" paddingHorizontal={6} paddingVertical={2} borderRadius="$2">
      Required
    </Text>
  )
}

export function InputPanel({ onSubmit, isLoading, ...frame }: InputPanelProps) {
  const [question, setQuestion] = useState('')
  const [context, setContext] = useState('')
  const [response, setResponse] = useState('')
  const [mode, setMode] = useState<'quick' | 'standard' | 'thorough'>('standard')
  const [domain, setDomain] = useState<'general' | 'medical' | 'legal' | 'financial'>('general')
  const [model, setModel] = useState<string>(AVAILABLE_MODELS[0])
  const [modelInputOpen, setModelInputOpen] = useState(false)
  const [modelSearch, setModelSearch] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = () => {
    if (!question.trim() || !context.trim() || !response.trim()) return
    onSubmit({ question, context, response, mode, domain, model })
  }

  const loadPreset = (preset: Preset) => {
    setQuestion(preset.question)
    setContext(preset.context)
    setResponse(preset.response)
    setMode(preset.mode || 'standard')
    setDomain(preset.domain || 'general')
  }

  const isValid = question.trim() && context.trim() && response.trim()
  const filteredModels = AVAILABLE_MODELS.filter((m) =>
    m.toLowerCase().includes(modelSearch.toLowerCase())
  )

  useEffect(() => {
    if (modelInputOpen && inputRef.current) inputRef.current.focus()
  }, [modelInputOpen])

  return (
    <YStack borderRightWidth={1} borderRightColor="$borderColor" backgroundColor="$color1" {...frame}>
      <YStack padding="$4" gap="$4" borderBottomWidth={1} borderBottomColor="$borderColor">
        <XStack alignItems="center" justifyContent="space-between">
          <Text fontSize={13} fontWeight="600" color="$color11" textTransform="uppercase">
            Configuration
          </Text>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" gap="$1">
                <Sparkles size={14} />
                <Text fontSize={13}>Load Existing</Text>
                <ChevronDown size={12} opacity={0.5} />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuLabel>Example Scenarios</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {PRESETS.map((preset) => (
                <DropdownMenuItem key={preset.id} onSelect={() => loadPreset(preset)}>
                  <YStack gap={2}>
                    <Text fontSize={14} fontWeight="500">
                      {preset.name}
                    </Text>
                    <Text fontSize={12} color="$color11">
                      {preset.description}
                    </Text>
                  </YStack>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </XStack>

        <XStack gap="$3">
          <YStack flex={1} gap={6}>
            <Label fontSize={12} color="$color11">
              Evaluation Mode
            </Label>
            <Select value={mode} onValueChange={(v: string) => setMode(v as typeof mode)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="quick" index={0}>
                  Quick
                </SelectItem>
                <SelectItem value="standard" index={1}>
                  Standard
                </SelectItem>
                <SelectItem value="thorough" index={2}>
                  Thorough
                </SelectItem>
              </SelectContent>
            </Select>
          </YStack>

          <YStack flex={1} gap={6}>
            <Label fontSize={12} color="$color11">
              Domain Context
            </Label>
            <Select value={domain} onValueChange={(v: string) => setDomain(v as typeof domain)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="general" index={0}>
                  General
                </SelectItem>
                <SelectItem value="medical" index={1}>
                  Medical
                </SelectItem>
                <SelectItem value="legal" index={2}>
                  Legal
                </SelectItem>
                <SelectItem value="financial" index={3}>
                  Financial
                </SelectItem>
              </SelectContent>
            </Select>
          </YStack>
        </XStack>

        <YStack gap={6}>
          <Label fontSize={12} color="$color11">
            Judge Model
          </Label>
          <Popover open={modelInputOpen} onOpenChange={setModelInputOpen}>
            <PopoverTrigger asChild>
              <Button variant="outline" justifyContent="space-between" width="100%">
                <Text fontFamily="$mono" fontSize={12} numberOfLines={1}>
                  {model}
                </Text>
                <ChevronDown size={12} opacity={0.5} />
              </Button>
            </PopoverTrigger>
            <PopoverContent padding={0} width={320}>
              <YStack padding="$2" borderBottomWidth={1} borderBottomColor="$borderColor">
                <Input
                  ref={inputRef}
                  placeholder="Search model..."
                  value={modelSearch}
                  onChangeText={setModelSearch}
                  fontFamily="$mono"
                  fontSize={12}
                  onKeyPress={(e: { nativeEvent: { key: string } }) => {
                    if (e.nativeEvent.key === 'Enter' && modelSearch.trim()) {
                      setModel(modelSearch.trim())
                      setModelSearch('')
                      setModelInputOpen(false)
                    }
                  }}
                />
              </YStack>
              <ScrollView maxHeight={192} padding="$1">
                {modelSearch.trim() && !knownModels.includes(modelSearch.trim()) && (
                  <XStack
                    alignItems="center"
                    gap="$2"
                    paddingHorizontal="$2"
                    paddingVertical={6}
                    borderRadius="$2"
                    cursor="pointer"
                    hoverStyle={{ backgroundColor: '$color3' }}
                    onPress={() => {
                      setModel(modelSearch.trim())
                      setModelSearch('')
                      setModelInputOpen(false)
                    }}
                  >
                    <Text fontSize={12} color="$color11">
                      Use:
                    </Text>
                    <Text fontSize={12} fontFamily="$mono" numberOfLines={1}>
                      {modelSearch.trim()}
                    </Text>
                  </XStack>
                )}
                {filteredModels.map((m) => (
                  <XStack
                    key={m}
                    alignItems="center"
                    gap="$2"
                    paddingHorizontal="$2"
                    paddingVertical={6}
                    borderRadius="$2"
                    cursor="pointer"
                    backgroundColor={model === m ? '$color3' : 'transparent'}
                    hoverStyle={{ backgroundColor: '$color3' }}
                    onPress={() => {
                      setModel(m)
                      setModelSearch('')
                      setModelInputOpen(false)
                    }}
                  >
                    {model === m ? <Check size={12} /> : <YStack width={12} />}
                    <Text fontSize={12} fontFamily="$mono" numberOfLines={1}>
                      {m}
                    </Text>
                  </XStack>
                ))}
              </ScrollView>
            </PopoverContent>
          </Popover>
        </YStack>
      </YStack>

      <ScrollView flex={1}>
        <YStack padding="$4" gap="$6">
          <YStack gap="$3">
            <XStack alignItems="center" justifyContent="space-between">
              <Label htmlFor="question" fontWeight="500">
                User Query
              </Label>
              <RequiredTag />
            </XStack>
            <Textarea
              id="question"
              value={question}
              onChangeText={setQuestion}
              placeholder="What question was asked?"
              minHeight={80}
              fontSize={14}
            />
          </YStack>

          <YStack gap="$3">
            <XStack alignItems="center" justifyContent="space-between">
              <Label htmlFor="context" fontWeight="500">
                Retrieved Context
              </Label>
              <RequiredTag />
            </XStack>
            <Textarea
              id="context"
              value={context}
              onChangeText={setContext}
              placeholder="Paste the source passages or context here..."
              minHeight={160}
              fontSize={14}
              fontFamily="$mono"
            />
          </YStack>

          <YStack gap="$3">
            <XStack alignItems="center" justifyContent="space-between">
              <Label htmlFor="response" fontWeight="500">
                Generated Response
              </Label>
              <RequiredTag />
            </XStack>
            <Textarea
              id="response"
              value={response}
              onChangeText={setResponse}
              placeholder="The answer generated by the system..."
              minHeight={120}
              fontSize={14}
            />
          </YStack>
        </YStack>
      </ScrollView>

      <YStack padding="$4" borderTopWidth={1} borderTopColor="$borderColor">
        <Button onPress={handleSubmit} disabled={isLoading || !isValid} size="lg" width="100%" gap="$2">
          {isLoading ? (
            <Text>Evaluating...</Text>
          ) : (
            <>
              <Play size={16} />
              <Text>Run Evaluation</Text>
            </>
          )}
        </Button>
      </YStack>
    </YStack>
  )
}
