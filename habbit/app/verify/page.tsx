import { redirect } from 'next/navigation';

export default async function Verify({ searchParams }: { searchParams: { token?: string } }) {
  const token = searchParams?.token;
  if (!token) return <div className="card">Verification failed.</div>;
  const res = await fetch(`${process.env.APP_URL}/api/auth/verify?token=${token}`, { cache: 'no-store' });
  if (!res.ok) return <div className="card">Verification failed.</div>;
  return <div className="card">Verified. You can login.</div>;
}
