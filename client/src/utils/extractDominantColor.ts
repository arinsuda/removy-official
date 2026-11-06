export interface ExtractedColor {
  r: number
  g: number
  b: number
  hex: string
  rgbStr: string
}

const DEFAULT_COLOR: ExtractedColor = {
  r: 229,
  g: 9,
  b: 20,
  hex: "#f5432c",
  rgbStr: "rgb(229, 9, 20)",
}

// In-memory cache to prevent re-extracting colors for the same image URL
const colorCache = new Map<string, ExtractedColor>()

/**
 * Extract dominant vibrant color from an image URL with in-memory caching and fast 20x30 canvas sampling.
 */
export async function extractDominantColor(
  imageSrc: string | null | undefined
): Promise<ExtractedColor> {
  if (!imageSrc) return DEFAULT_COLOR

  if (colorCache.has(imageSrc)) {
    return colorCache.get(imageSrc)!
  }

  return new Promise((resolve) => {
    const img = new Image()
    img.crossOrigin = "Anonymous"

    img.onload = () => {
      try {
        const canvas = document.createElement("canvas")
        const ctx = canvas.getContext("2d", { willReadFrequently: true })
        if (!ctx) return resolve(DEFAULT_COLOR)

        // Fast tiny canvas sampling (20x30 = 600 pixels)
        const width = 20
        const height = 30
        canvas.width = width
        canvas.height = height

        ctx.drawImage(img, 0, 0, width, height)

        const imageData = ctx.getImageData(0, 0, width, height)
        const data = imageData.data

        let totalR = 0
        let totalG = 0
        let totalB = 0
        let count = 0

        let vibrantR = 0
        let vibrantG = 0
        let vibrantB = 0
        let maxScore = -1

        for (let i = 0; i < data.length; i += 4) {
          const r = data[i] ?? 0
          const g = data[i + 1] ?? 0
          const b = data[i + 2] ?? 0
          const a = data[i + 3] ?? 255

          if (a < 128) continue

          const max = Math.max(r, g, b)
          const min = Math.min(r, g, b)
          const brightness = (r + g + b) / 3
          const saturation = max === 0 ? 0 : (max - min) / max

          if (brightness < 20 || brightness > 240) continue

          totalR += r
          totalG += g
          totalB += b
          count++

          const score = saturation * 2 + (1 - Math.abs(brightness - 128) / 128)
          if (score > maxScore) {
            maxScore = score
            vibrantR = r
            vibrantG = g
            vibrantB = b
          }
        }

        if (count === 0 && maxScore < 0) {
          colorCache.set(imageSrc, DEFAULT_COLOR)
          return resolve(DEFAULT_COLOR)
        }

        const r = maxScore > 0.5 ? vibrantR : Math.round(totalR / (count || 1))
        const g = maxScore > 0.5 ? vibrantG : Math.round(totalG / (count || 1))
        const b = maxScore > 0.5 ? vibrantB : Math.round(totalB / (count || 1))

        const hex = `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`
        const result: ExtractedColor = {
          r,
          g,
          b,
          hex,
          rgbStr: `rgb(${r}, ${g}, ${b})`,
        }

        colorCache.set(imageSrc, result)
        resolve(result)
      } catch {
        resolve(DEFAULT_COLOR)
      }
    }

    img.onerror = () => {
      resolve(DEFAULT_COLOR)
    }

    img.src = imageSrc
  })
}
