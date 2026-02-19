import '../styles/global.css';
import React from 'react';

export const metadata = {
  title: 'Habbit',
  description: 'Habit tracker'
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="container">
          {children}
        </div>
      </body>
    </html>
  );
}
