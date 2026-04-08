import { withRoute } from "@/lib/with-route";
import { prisma } from "@/lib/db";

export const GET = withRoute("/nextjs/api/users", async () => {
  const users = await prisma.user.findMany();
  return Response.json(users);
});

export const POST = withRoute("/nextjs/api/users", async (req) => {
  const body = await req.json();
  const user = await prisma.user.create({
    data: { name: body.name, email: body.email },
  });
  return Response.json(user, { status: 201 });
});
