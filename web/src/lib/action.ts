/** Run a promise and route rejection to the page error slot instead of leaving it unhandled. */
export function run(promise: Promise<unknown>, onError: (message: string) => void): void {
  void promise.catch((error: unknown) => {
    onError(error instanceof Error ? error.message : String(error))
  })
}
