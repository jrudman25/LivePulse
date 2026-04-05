import { render, screen, fireEvent } from '@testing-library/react';
import EventFeed from './EventFeed';
import React from 'react';

// Mocking EventCard so we don't accidentally execute background React API fetch hooks inside testing environments
jest.mock('./EventCard', () => {
  return function MockEventCard({ event }: { event: any }) {
    return <div data-testid="event-card">{event.title}</div>;
  };
});

const mockEvents = [
  { id: '1', type: 'Music', title: 'Coachella', country: 'US', location: 'California' },
  { id: '2', type: 'Sports', title: 'World Cup', country: 'CA', location: 'Toronto' }
];

describe('EventFeed Filters & Layouts', () => {
  it('renders all provided events perfectly on initial mount', () => {
    render(<EventFeed initialEvents={mockEvents} />);
    
    // Validate both cards render natively
    const cards = screen.getAllByTestId('event-card');
    expect(cards).toHaveLength(2);
    expect(screen.getByText('Coachella')).toBeInTheDocument();
    expect(screen.getByText('World Cup')).toBeInTheDocument();
  });

  it('filters the feed automatically when the user selects a specific Event Type', () => {
    render(<EventFeed initialEvents={mockEvents} />);
    
    // There are two Dropdowns. 
    // The Event Type dropdown starts rendering "All". The Country dropdown renders "All Countries".
    const typeDropdown = screen.getByDisplayValue('All');
    
    // Simulate user selecting "Music" exclusively
    fireEvent.change(typeDropdown, { target: { value: 'Music' } });
    
    // World Cup should gracefully unmount from DOM
    expect(screen.queryByText('World Cup')).not.toBeInTheDocument();
    expect(screen.getByText('Coachella')).toBeInTheDocument();
  });

  it('filters explicitly based on Country', () => {
    render(<EventFeed initialEvents={mockEvents} />);
    
    const countryDropdown = screen.getByDisplayValue('All Countries');
    
    // Simulate user toggling directly mapping to `CA`
    fireEvent.change(countryDropdown, { target: { value: 'CA' } });
    
    // World Cup strictly renders since it's located in CA. Coachella is dropped.
    expect(screen.getByText('World Cup')).toBeInTheDocument();
    expect(screen.queryByText('Coachella')).not.toBeInTheDocument();
  });

  it('gracefully renders empty fallback state when no tickets perfectly align with filter constraints', () => {
    render(<EventFeed initialEvents={mockEvents} />);
    
    const typeDropdown = screen.getByDisplayValue('All');
    const countryDropdown = screen.getByDisplayValue('All Countries');
    
    fireEvent.change(typeDropdown, { target: { value: 'Music' } }); // filters US, leaves Coachella
    fireEvent.change(countryDropdown, { target: { value: 'CA' } }); // further limits array searching CA

    expect(screen.queryByTestId('event-card')).not.toBeInTheDocument();
    expect(screen.getByText(/No events found/i)).toBeInTheDocument();
  });
});
