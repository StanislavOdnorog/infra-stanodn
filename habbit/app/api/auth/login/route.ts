import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';
import bcrypt from 'bcryptjs';
import { createSession, normalizeEmail } from '../../../../lib/auth';

export async function POST(req: Request) {
  const { email, password } = await req.json();
  const normalized = normalizeEmail(email);
  const user = await prisma.user.findUnique({ where: { email: normalized } });
  if (!user) return NextResponse.json({ error: 'Invalid' }, { status: 401 });
  const ok = await bcrypt.compare(password, user.passwordHash);
  if (!ok) return NextResponse.json({ error: 'Invalid' }, { status: 401 });
  if (!user.emailVerified) return NextResponse.json({ error: 'Not verified' }, { status: 403 });
  await createSession(user.id);
  return NextResponse.json({ ok: true });
}
