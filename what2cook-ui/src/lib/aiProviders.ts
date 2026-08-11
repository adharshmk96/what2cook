export type AiProviderId = 'chatgpt' | 'claude' | 'gemini' | 'grok'

export type AiProvider = {
  id: AiProviderId
  label: string
  buildUrl: (prompt: string) => string
}

export const AI_PROVIDERS: AiProvider[] = [
  {
    id: 'chatgpt',
    label: 'ChatGPT',
    buildUrl: (prompt) => `https://chatgpt.com/?q=${encodeURIComponent(prompt)}`,
  },
  {
    id: 'claude',
    label: 'Claude',
    buildUrl: (prompt) => `https://claude.ai/new?q=${encodeURIComponent(prompt)}`,
  },
  {
    id: 'gemini',
    label: 'Gemini',
    buildUrl: (prompt) =>
      `https://gemini.google.com/app?q=${encodeURIComponent(prompt)}`,
  },
  {
    id: 'grok',
    label: 'Grok',
    buildUrl: (prompt) => `https://grok.com/?q=${encodeURIComponent(prompt)}`,
  },
]

export function openAiProvider(provider: AiProvider, prompt: string): void {
  const trimmed = prompt.trim()
  if (!trimmed) {
    return
  }
  window.open(provider.buildUrl(trimmed), '_blank', 'noopener,noreferrer')
}
