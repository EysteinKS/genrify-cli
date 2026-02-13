export class CancelledError extends Error {
  constructor(message = 'Cancelled') {
    super(message)
    this.name = 'CancelledError'
  }
}

export function isCancelledError(err: unknown): boolean {
  return err instanceof CancelledError || (err instanceof Error && err.name === 'CancelledError')
}
