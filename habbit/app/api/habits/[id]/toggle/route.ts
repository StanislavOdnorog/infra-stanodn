import { NextResponse } from 'next/server';
import { getCurrentUser } from '../../../../../lib/auth';
import { prisma } from '../../../../../lib/db';

export async function POST(req: Request, { params }: { params: { id: string } }) {
  const user = await getCurrentUser();
  if (!user) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  const habit = await prisma.habit.findFirst({ where: { id: params.id, userId: user.id } });
  if (!habit) return NextResponse.json({ error: 'Not found' }, { status: 404 });
  const today = new Date();
  today.setHours(0,0,0,0);
  const existing = await prisma.habitLog.findUnique({ where: { habitId_date: { habitId: habit.id, date: today } } });
  if (existing) {
    await prisma.habitLog.delete({ where: { id: existing.id } });
  } else {
    await prisma.habitLog.create({ data: { habitId: habit.id, date: today } });
  }
  return NextResponse.redirect(new URL('/dashboard', req.url));
}
