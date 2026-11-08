/** Number of kernels on the meter; each stands for half a point of the 0-5 POPCORN score. */
export const KERNEL_COUNT = 10

/** Fill (0 to 1) of each kernel for a 0-5 score, e.g. 3.7 fills 7 kernels and 40% of the 8th. */
export function kernelFills(average: number | null | undefined): number[] {
  const halves = average && Number.isFinite(average) ? Math.min(5, Math.max(0, average)) * 2 : 0
  return Array.from({ length: KERNEL_COUNT }, (_, i) => Math.min(1, Math.max(0, halves - i)))
}
