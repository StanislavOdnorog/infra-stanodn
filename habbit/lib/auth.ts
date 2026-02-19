import { cookies } from 'next/headers';
import jwt from 'jsonwebtoken';
import { prisma } from './db';

const SESSION_COOKIE = 'habbit_session';
const SESSION_TTL_DAYS = 30;

export async function createSession(userId: string) {
  const token = jwt.sign({ sub: userId }, process.env.JWT_SECRET!, {
    expiresIn: `${SESSION_TTL_DAYS}d`
  });
  const expiresAt = new Date(Date.now() + SESSION_TTL_DAYS * 24 * 60 * 60 * 1000);
  await prisma.session.create({ data: { userId, token, expiresAt } });
  cookies().set(SESSION_COOKIE, token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    expires: expiresAt,
    path: '/'
  });
}

export function normalizeEmail(email: string) {
  return email.trim().toLowerCase();
}

export async function getCurrentUser() {
  const token = cookies().get(SESSION_COOKIE)?.value;
  if (!token) return null;
  try {
    const payload = jwt.verify(token, process.env.JWT_SECRET!) as { sub: string };
    const session = await prisma.session.findFirst({ where: { token } });
    if (!session || session.expiresAt < new Date()) return null;
    return prisma.user.findUnique({ where: { id: payload.sub } });
  } catch {
    return null;
  }
}

export async function clearSession() {
  const token = cookies().get(SESSION_COOKIE)?.value;
  if (token) {
    await prisma.session.deleteMany({ where: { token } });
  }
  cookies().set(SESSION_COOKIE, '', { expires: new Date(0), path: '/', httpOnly: true, sameSite: 'lax', secure: process.env.NODE_ENV === 'production' });
}
