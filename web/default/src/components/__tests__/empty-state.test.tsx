import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { EmptyState } from '../empty-state'

describe('EmptyState', () => {
  it('renders with default props', () => {
    render(<EmptyState title="No data" />)
    expect(screen.getByText('No data')).toBeDefined()
  })

  it('renders description when provided', () => {
    render(<EmptyState title="No data" description="Nothing here yet" />)
    expect(screen.getByText('Nothing here yet')).toBeDefined()
  })

  it('renders action button when provided', () => {
    const onClick = vi.fn()
    render(<EmptyState title="No data" action={{ label: 'Add Item', onClick }} />)
    const button = screen.getByText('Add Item')
    expect(button).toBeDefined()
  })

  it('calls onClick when action button is clicked', () => {
    const onClick = vi.fn()
    render(<EmptyState title="No data" action={{ label: 'Add Item', onClick }} />)
    fireEvent.click(screen.getByText('Add Item'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders without description by default', () => {
    render(<EmptyState title="No data" />)
    expect(screen.queryByText('Nothing here yet')).toBeNull()
  })

  it('does not render action button when not provided', () => {
    render(<EmptyState title="No data" />)
    expect(screen.queryByRole('button')).toBeNull()
  })
})
