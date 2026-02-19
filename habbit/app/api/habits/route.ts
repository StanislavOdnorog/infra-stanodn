import { NextResponse } from 'next/server';
import { getCurrentUser } from '../../../lib/auth';
import { prisma } from '../../../lib/db';

export async function GET() {
  const user = await getCurrentUser();
  if (!user) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  const habits = await prisma.habit.findMany({ where: { userId: user.id }, include: { logs: true } });
  return NextResponse.json(habits);
}

export async function POST(req: Request) {
  const user = await getCurrentUser();
  if (!user) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  const contentType = req.headers.get('content-type') || '';
  let body: any = {};
  if (contentType.includes('application/json')) {
    body = await req.json();
  } else if (contentType.includes('multipart/form-data') || contentType.includes('application/x-www-form-urlencoded')) {
    const form = await req.formData();
    body.title = String(form.get('title') || '');
  }
  if (!body.title) return NextResponse.json({ error: 'Title required' }, { status: 400 });
  const habit = await prisma.habit.create({
    data: {
      userId: user.id,
      title: body.title,
      description: body.description,
      frequency: body.frequency || 'DAILY',
      daysOfWeek: body.daysOfWeek || [],
      timesPerWeek: body.timesPerWeek,
      everyXDays: body.everyXDays
    }
  });
  if (contentType.includes('application/json')) return NextResponse.json(habit);
  return NextResponse.redirect(new URL('/dashboard', req.url));
}
