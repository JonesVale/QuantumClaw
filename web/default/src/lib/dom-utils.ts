export function applyFaviconToDom(logoUrl: string): void {
  if (typeof document === 'undefined') return
  let link = document.querySelector<HTMLLinkElement>(
    "link[rel~='icon']",
  )
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.href = logoUrl
}
