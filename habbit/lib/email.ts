import nodemailer from 'nodemailer';

export function getTransport() {
  return nodemailer.createTransport({
    host: process.env.SMTP_HOST,
    port: Number(process.env.SMTP_PORT || 465),
    secure: true,
    auth: {
      user: process.env.SMTP_USER,
      pass: process.env.SMTP_PASS
    }
  });
}

export async function sendVerificationEmail(to: string, token: string, baseUrl: string) {
  const url = `${baseUrl}/verify?token=${token}`;
  const transporter = getTransport();
  await transporter.sendMail({
    from: process.env.SMTP_FROM,
    to,
    subject: 'Verify your email',
    text: `Verify your email: ${url}`
  });
}
