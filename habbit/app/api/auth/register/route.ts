import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';
import bcrypt from 'bcryptjs';
import { sendVerificationEmail } from '../../../../lib/email';
import { randomBytes } from 'crypto';

export async function POST(req: Request) {
  const { email, password } = await req.json();
  if (!email || !password) return NextResponse.json({ error: 'Invalid' }, { status: 400 });
  const existing = await prisma.user.findUnique({ where: { email } });
  if (existing) return NextResponse.json({ error: 'Exists' }, { status: 400 });
  const passwordHash = await bcrypt.hash(password, 10);
  const user = await prisma.user.create({ data: { email, passwordHash } });
  const token = randomBytes(32).toString('hex');
  await prisma.token.create({ data: { userId: user.id, token, type: 'verify', expiresAt: new Date(Date.now()+1000*60*60*24) } });
  await sendVerificationEmail(email, token);
  return NextResponse.json({ ok: true });
}
