import { RegisterForm } from "@/features/auth";

export default function RegisterPage() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-6 p-8">
      <h1 className="text-2xl font-semibold">Create account</h1>
      <RegisterForm />
    </div>
  );
}
