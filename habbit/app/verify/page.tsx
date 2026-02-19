import { verifyEmailToken } from '../../lib/verify';

export default async function Verify({ searchParams }: { searchParams: { token?: string } }) {
  const token = searchParams?.token;
  if (!token) return <div className="card">Verification failed.</div>;
  const ok = await verifyEmailToken(token);
  return <div className="card">{ok ? 'Verified. You can login.' : 'Verification failed.'}</div>;
}
