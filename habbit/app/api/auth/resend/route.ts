import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';
import { randomBytes, createHash } from 'crypto';
import { sendVerificationEmail } from '../../../../lib/email';
import { normalizeEmail } from '../../../../lib/auth';

export async function POST(req: Request) {
  const { email } = await req.json();
  if (!email) return NextResponse.json({ error: 'Invalid input' }, { status: 400 });
  const normalized = normalizeEmail(email);
  const user = await prisma.user.findUnique({ where: { email: normalized } });
  if (!user) return NextResponse.json({ ok: true });
  const token = randomBytes(32).toString('hex');
  const hash = createHash('sha256').update(token).digest('hex');
  await prisma.token.create({ data: { userId: user.id, token: hash, type: 'verify', expiresAt: new Date(Date.now()+1000*60*60*24) } });
  const baseUrl = process.env.APP_URL || new URL(req.url).origin;
  await sendVerificationEmail(normalized, token, baseUrl);
  return NextResponse.json({ ok: true });
}
