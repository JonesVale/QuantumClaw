import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from '../error-boundary'

describe('ErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div>test content</div>
      </ErrorBoundary>
    )
    expect(screen.getByText('test content')).toBeDefined()
  })

  it('renders error fallback on error', () => {
    const originalError = console.error
    console.error = () => {}

    const ThrowError = () => {
      throw new Error('test error')
    }

    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    )
    expect(screen.getByText(/something went wrong/i)).toBeDefined()

    console.error = originalError
  })

  it('renders custom fallback when provided', () => {
    const originalError = console.error
    console.error = () => {}

    const ThrowError = () => {
      throw new Error('test error')
    }

    render(
      <ErrorBoundary fallback={<div>Custom error UI</div>}>
        <ThrowError />
      </ErrorBoundary>
    )
    expect(screen.getByText('Custom error UI')).toBeDefined()

    console.error = originalError
  })

  it('displays the error message', () => {
    const originalError = console.error
    console.error = () => {}

    const ThrowError = () => {
      throw new Error('Something broke!')
    }

    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    )
    expect(screen.getByText('Something broke!')).toBeDefined()

    console.error = originalError
  })
})
