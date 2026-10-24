import type { PointerEvent as ReactPointerEvent } from "react"
import { useEffect, useLayoutEffect, useRef, useState } from "react"

type DropPosition = "before" | "after"
type DropTarget = { targetKey: string; position: DropPosition }

interface PointerDragSession {
  sourceKey: string
  pointerId: number
  sourceRow: HTMLElement
  startX: number
  startY: number
  offsetX: number
  offsetY: number
  started: boolean
}

interface PointerDragImage {
  move: (clientX: number, clientY: number) => void
  remove: () => void
}

function createPointerDragImage(
  source: HTMLElement,
  offsetX: number,
  offsetY: number
): PointerDragImage {
  const rect = source.getBoundingClientRect()
  const image = source.cloneNode(true) as HTMLElement
  image.setAttribute("aria-hidden", "true")
  image.setAttribute("inert", "")
  image.removeAttribute("id")
  image.removeAttribute("data-flip-id")
  image.removeAttribute("data-sort-key")
  image.draggable = false
  image.querySelectorAll("[id]").forEach((element) => {
    element.removeAttribute("id")
  })
  image.querySelectorAll("input").forEach((input, index) => {
    const sourceInput = source.querySelectorAll("input")[index]
    if (sourceInput) {
      input.value = sourceInput.value
    }
  })
  Object.assign(image.style, {
    position: "fixed",
    top: "0",
    left: "0",
    width: `${rect.width}px`,
    height: `${rect.height}px`,
    margin: "0",
    pointerEvents: "none",
    zIndex: "2147483647",
    opacity: "0.96",
    willChange: "transform",
  })
  document.body.appendChild(image)

  return {
    move(clientX, clientY) {
      image.style.transform = `translate3d(${clientX - offsetX}px, ${clientY - offsetY}px, 0)`
    },
    remove() {
      image.remove()
    },
  }
}

/**
 * Pointer-based sorting keeps the dragged visual independent of the source row,
 * so the source can leave layout flow without cancelling a native drag session.
 */
export function usePointerReorder<T extends HTMLElement>(
  revision: unknown,
  axis: "x" | "y",
  onDrop: (sourceKey: string, targetKey: string, position: DropPosition) => void
) {
  const [dragKey, setDragKey] = useState("")
  const [dropTarget, setDropTarget] = useState("")
  const [dropPosition, setDropPosition] = useState<DropPosition>("after")
  const [dragSize, setDragSize] = useState({ width: 0, height: 0 })
  const flipRevision = JSON.stringify([
    revision,
    dragKey,
    dropTarget,
    dropPosition,
  ])
  const { containerRef, capturePositions } = useFlipReorder<T>(flipRevision)
  const sessionRef = useRef<PointerDragSession | null>(null)
  const dropRef = useRef<DropTarget>({ targetKey: "", position: "after" })
  const dragImageRef = useRef<PointerDragImage | null>(null)
  const listenerCleanupRef = useRef<(() => void) | null>(null)
  const onDropRef = useRef(onDrop)
  const suppressClickRef = useRef("")

  useLayoutEffect(() => {
    onDropRef.current = onDrop
  }, [onDrop])

  function updateDropTarget(
    session: PointerDragSession,
    clientX: number,
    clientY: number
  ) {
    const container = containerRef.current
    const hovered = document.elementFromPoint(clientX, clientY)
    let next: DropTarget = { targetKey: "", position: "after" }

    if (container && hovered && container.contains(hovered)) {
      const slot = hovered.closest<HTMLElement>("[data-drop-slot]")
      if (slot) {
        const targetKey = slot.dataset.dropTarget ?? ""
        const position = slot.dataset.dropPosition
        if (targetKey && (position === "before" || position === "after")) {
          next = { targetKey, position }
        }
      } else {
        const row = hovered.closest<HTMLElement>("[data-sort-key]")
        const targetKey = row?.dataset.sortKey ?? ""
        if (row && targetKey && targetKey !== session.sourceKey) {
          const rect = row.getBoundingClientRect()
          const coordinate = axis === "x" ? clientX : clientY
          const midpoint =
            axis === "x"
              ? rect.left + rect.width / 2
              : rect.top + rect.height / 2
          next = {
            targetKey,
            position: coordinate < midpoint ? "before" : "after",
          }
        }
      }
    }

    const previous = dropRef.current
    if (
      previous.targetKey !== next.targetKey ||
      previous.position !== next.position
    ) {
      capturePositions()
      dropRef.current = next
      setDropTarget(next.targetKey)
      setDropPosition(next.position)
    }
  }

  function startDrag(sourceKey: string, event: ReactPointerEvent<Element>) {
    if (event.button !== 0 || event.isPrimary === false || sessionRef.current) {
      return
    }
    const sourceRow =
      event.currentTarget.closest<HTMLElement>("[data-sort-key]")
    const container = containerRef.current
    if (!sourceRow || !container?.contains(sourceRow)) {
      return
    }

    const rect = sourceRow.getBoundingClientRect()
    sessionRef.current = {
      sourceKey,
      pointerId: event.pointerId,
      sourceRow,
      startX: event.clientX,
      startY: event.clientY,
      offsetX: event.clientX - rect.left,
      offsetY: event.clientY - rect.top,
      started: false,
    }
    dropRef.current = { targetKey: "", position: "after" }

    const onMove = (pointerEvent: PointerEvent) => onPointerMove(pointerEvent)
    const onUp = (pointerEvent: PointerEvent) =>
      finishPointerDrag(pointerEvent, true)
    const onCancel = (pointerEvent: PointerEvent) =>
      finishPointerDrag(pointerEvent, false)
    window.addEventListener("pointermove", onMove, {
      capture: true,
      passive: false,
    })
    window.addEventListener("pointerup", onUp, true)
    window.addEventListener("pointercancel", onCancel, true)
    listenerCleanupRef.current = () => {
      window.removeEventListener("pointermove", onMove, true)
      window.removeEventListener("pointerup", onUp, true)
      window.removeEventListener("pointercancel", onCancel, true)
    }
  }

  function onPointerMove(event: PointerEvent) {
    const session = sessionRef.current
    if (!session || session.pointerId !== event.pointerId) {
      return
    }
    const distance = Math.hypot(
      event.clientX - session.startX,
      event.clientY - session.startY
    )
    if (!session.started) {
      if (distance < 5) {
        return
      }
      session.started = true
      event.preventDefault()
      capturePositions()
      const rect = session.sourceRow.getBoundingClientRect()
      dragImageRef.current = createPointerDragImage(
        session.sourceRow,
        session.offsetX,
        session.offsetY
      )
      dragImageRef.current.move(event.clientX, event.clientY)
      setDragSize({ width: rect.width, height: rect.height })
      setDragKey(session.sourceKey)
    } else {
      event.preventDefault()
      dragImageRef.current?.move(event.clientX, event.clientY)
    }
    updateDropTarget(session, event.clientX, event.clientY)
  }

  function finishPointerDrag(event: PointerEvent, shouldCommit: boolean) {
    const session = sessionRef.current
    if (!session || session.pointerId !== event.pointerId) {
      return
    }
    if (session.started) {
      event.preventDefault()
      if (shouldCommit) {
        updateDropTarget(session, event.clientX, event.clientY)
      }
      capturePositions()
      const drop = dropRef.current
      sessionRef.current = null
      if (
        shouldCommit &&
        drop.targetKey &&
        drop.targetKey !== session.sourceKey
      ) {
        onDropRef.current(session.sourceKey, drop.targetKey, drop.position)
      }
      suppressClickRef.current = session.sourceKey
      window.setTimeout(() => {
        if (suppressClickRef.current === session.sourceKey) {
          suppressClickRef.current = ""
        }
      }, 0)
      setDragKey("")
      setDropTarget("")
      setDropPosition("after")
      setDragSize({ width: 0, height: 0 })
      dragImageRef.current?.remove()
      dragImageRef.current = null
    }
    sessionRef.current = null
    listenerCleanupRef.current?.()
    listenerCleanupRef.current = null
  }

  function consumeClick(sourceKey: string) {
    if (suppressClickRef.current !== sourceKey) {
      return false
    }
    suppressClickRef.current = ""
    return true
  }

  useEffect(
    () => () => {
      sessionRef.current = null
      listenerCleanupRef.current?.()
      listenerCleanupRef.current = null
      dragImageRef.current?.remove()
      dragImageRef.current = null
    },
    []
  )

  return {
    containerRef,
    dragKey,
    dropTarget,
    dropPosition,
    dragSize,
    startDrag,
    consumeClick,
    capturePositions,
  }
}

/**
 * Captures the current positions of keyed children and animates their layout
 * changes with transforms after `revision` changes (a small FLIP transition).
 */
export function useFlipReorder<T extends HTMLElement>(revision: unknown) {
  const containerRef = useRef<T | null>(null)
  const beforeRef = useRef<Map<string, DOMRect> | null>(null)
  const framesRef = useRef<number[]>([])
  const animatedElementsRef = useRef<HTMLElement[]>([])

  function capturePositions() {
    const container = containerRef.current
    if (!container) {
      return
    }

    const positions = new Map<string, DOMRect>()
    container
      .querySelectorAll<HTMLElement>("[data-flip-id]")
      .forEach((element) => {
        const id = element.dataset.flipId
        if (id) {
          positions.set(id, element.getBoundingClientRect())
        }
      })
    beforeRef.current = positions
  }

  useLayoutEffect(() => {
    const before = beforeRef.current
    const container = containerRef.current
    beforeRef.current = null
    if (!before || !container) {
      return
    }

    const reducedMotion =
      typeof window !== "undefined" &&
      window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
    const elements = Array.from(
      container.querySelectorAll<HTMLElement>("[data-flip-id]")
    )

    // Read every final position before writing any animation styles.
    const changes = elements.flatMap((element) => {
      const id = element.dataset.flipId
      const first = id ? before.get(id) : undefined
      if (!first) {
        return []
      }
      const last = element.getBoundingClientRect()
      const x = first.left - last.left
      const y = first.top - last.top
      return x || y ? [{ element, x, y }] : []
    })

    animatedElementsRef.current = changes.map(({ element }) => element)
    if (reducedMotion || changes.length === 0) {
      changes.forEach(({ element }) => {
        element.style.transition = ""
        element.style.transform = ""
      })
      return
    }

    changes.forEach(({ element, x, y }) => {
      element.style.transition = "none"
      element.style.transform = `translate3d(${x}px, ${y}px, 0)`
    })
    // Flush the inverse transforms once, after all writes, before animating to
    // the new layout. This avoids per-item read/write layout thrashing.
    void container.getBoundingClientRect()

    const frame = window.requestAnimationFrame(() => {
      changes.forEach(({ element }) => {
        element.style.transition = "transform 180ms cubic-bezier(0.2, 0, 0, 1)"
        element.style.transform = "translate3d(0, 0, 0)"
      })
    })
    framesRef.current.push(frame)
  }, [revision])

  useLayoutEffect(
    () => () => {
      framesRef.current.forEach((frame) => window.cancelAnimationFrame(frame))
      animatedElementsRef.current.forEach((element) => {
        element.style.transition = ""
        element.style.transform = ""
      })
    },
    []
  )

  return { containerRef, capturePositions }
}
