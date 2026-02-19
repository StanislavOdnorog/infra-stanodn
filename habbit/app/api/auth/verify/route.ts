import { NextResponse } from 'next/server';
import { prisma } from '../../../../lib/db';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const token = searchParams.get('token');
  if (!token) return NextResponse.json({ error: 'Invalid' }, { status: 400 });
  const record = await prisma.token.findUnique({ where: { token } });
  if (!record || record.expiresAt < new Date()) return NextResponse.json({ error: 'Expired' }, { status: 400 });
  await prisma.user.update({ where: { id: record.userId }, data: { emailVerified: true } });
  await prisma.token.delete({ where: { id: record.id } });
  return NextResponse.json({ ok: true });
}
