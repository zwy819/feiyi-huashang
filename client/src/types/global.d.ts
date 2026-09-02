export {}

declare global {
  interface Window {
    __showToast?: (message: string, type?: string) => void
  }
}
