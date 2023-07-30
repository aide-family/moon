import React, { useRef, useEffect } from 'react'

export interface PromQLInputProps {
  value?: string
  onChange?: (value: string) => void
}

const PromQLInput: React.FC<PromQLInputProps> = ({ value, onChange }) => {
  const editorRef = useRef<HTMLDivElement>(null)

  return <div ref={editorRef} />
}

export default PromQLInput
