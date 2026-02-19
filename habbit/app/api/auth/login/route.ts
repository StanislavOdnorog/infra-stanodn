import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';
import bcrypt from 'bcryptjs';
import { createSession } from '../../../../lib/auth';

export async function POST(req: Request) {
  const { email, password } = await req.json();
  const user = await prisma.user.findUnique({ where: { email } });
  if (!user) return NextResponse.json({ error: 'Invalid' }, { status: 401 });
  const ok = await bcrypt.compare(password, user.passwordHash);
  if (!ok || !user.emailVerified) return NextResponse.json({ error: 'Invalid' }, { status: 401 });
  await createSession(user.id);
  return NextResponse.json({ ok: true });
}
