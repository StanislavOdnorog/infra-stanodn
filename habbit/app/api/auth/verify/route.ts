import { NextResponse } from 'next/server';
import { verifyEmailToken } from '../../../../lib/verify';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const token = searchParams.get('token');
  if (!token) return NextResponse.json({ error: 'Invalid' }, { status: 400 });
  const ok = await verifyEmailToken(token);
  if (!ok) return NextResponse.json({ error: 'Expired' }, { status: 400 });
  return NextResponse.json({ ok: true });
}
