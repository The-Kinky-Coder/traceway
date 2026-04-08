import { withRoute } from "@/lib/with-route";
import { prisma } from "@/lib/db";

export const GET = withRoute(
  "/nextjs/api/users/[id]",
  async (req, { params }) => {
    const { id } = await params;
    const user = await prisma.user.findUnique({
      where: { id: parseInt(id) },
    });
    if (!user) {
      return Response.json({ error: "User not found" }, { status: 404 });
    }
    return Response.json(user);
  }
);
