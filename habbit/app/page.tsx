import Link from 'next/link';
import { getCurrentUser } from '../lib/auth';
import { redirect } from 'next/navigation';

export default async function Home() {
  const user = await getCurrentUser();
  if (user) redirect('/dashboard');
  return (
    <div className="grid">
      <h1>Habbit</h1>
      <p>Track habits with flexible schedules and streaks.</p>
      <div className="row">
        <Link href="/auth/login">Login</Link>
        <Link href="/auth/register">Register</Link>
      </div>
    </div>
  );
}
