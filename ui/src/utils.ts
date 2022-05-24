export function numberToRgba(n: number): string {
  const r = (n >> 24) & 0xff;
  const g = (n >> 16) & 0xff;
  const b = (n >> 8) & 0xff;
  const a = n & 0xff;
  return `rgba(${r},${g},${b},${a})`;
}
