import { prisma } from './db';
import { createHash } from 'crypto';

export async function verifyEmailToken(token: string) {
  const hash = createHash('sha256').update(token).digest('hex');
  const record = await prisma.token.findFirst({ where: { token: hash, type: 'verify' } });
  if (!record || record.expiresAt < new Date()) return false;
  await prisma.user.update({ where: { id: record.userId }, data: { emailVerified: true } });
  await prisma.token.delete({ where: { id: record.id } });
  return true;
}
