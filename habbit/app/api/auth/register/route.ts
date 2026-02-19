import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';
import bcrypt from 'bcryptjs';
import { sendVerificationEmail } from '../../../../lib/email';
import { randomBytes, createHash } from 'crypto';
import { normalizeEmail } from '../../../../lib/auth';

export async function POST(req: Request) {
  const { email, password } = await req.json();
  if (!email || !password || password.length < 8) {
    return NextResponse.json({ error: 'Invalid input' }, { status: 400 });
  }
  const normalized = normalizeEmail(email);
  const existing = await prisma.user.findUnique({ where: { email: normalized } });
  if (existing) return NextResponse.json({ error: 'Email already exists' }, { status: 400 });
  const passwordHash = await bcrypt.hash(password, 10);
  const user = await prisma.user.create({ data: { email: normalized, passwordHash } });
  const token = randomBytes(32).toString('hex');
  const hash = createHash('sha256').update(token).digest('hex');
  await prisma.token.create({ data: { userId: user.id, token: hash, type: 'verify', expiresAt: new Date(Date.now()+1000*60*60*24) } });
  const baseUrl = process.env.APP_URL || new URL(req.url).origin;
  try {
    await sendVerificationEmail(normalized, token, baseUrl);
  } catch (err) {
    return NextResponse.json({ error: 'Email send failed' }, { status: 500 });
  }
  return NextResponse.json({ ok: true });
}
